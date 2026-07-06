# Go Asteroids — Refactor Plan

Goal: break the ~24-file `goasteroids` god-package into clear, one-directional
packages (directories) without breaking the game.

## Two facts that drive everything

1. **In Go, a directory *is* a package.** "Break into directories" = "break into
   packages", which means breaking the current coupling — you can't just move
   files into subfolders and keep `package goasteroids`.
2. **The expensive coupling is field-level reach-in.** `GameScene`'s
   collision/spawn methods in `game-scene.go` read and write entities'
   *unexported* fields directly (`m.sprite`, `l.laserObj`, and ~13 fields on
   `Player`). Splitting entities into their own package forces every one of
   those to go through an exported field or an accessor method. This — not the
   `game *GameScene` back-reference — is the bulk of the work.

### Coupling map (why a phased split is possible)

| Entity | Coupling to scene | Split difficulty |
|---|---|---|
| `Star`, `Exhaust`, `AlienLaser`, `LifeIndicator`, `ShieldIndicator`, `HyperspaceIndicator` | **None** — pure Update/Draw, no `game` field | Trivial |
| `Meteor`, `Alien`, `Laser` | Store `game` but their own methods barely use it; scene reads their fields | Easy–moderate |
| `Shield` | Needs `space` + player geometry | Moderate |
| `Player` | Deeply intertwined with `GameScene` both directions | Hard |

---

## Target architecture

```
go-asteroids/
├── main.go
├── assets/                      # leaf — already clean
└── internal/
    ├── engine/                  # LEAF: Vector, Timer, Tag, ObjectData, config consts, checkCollision
    ├── highscore/               # LEAF: the file I/O from helpers.go
    ├── entity/                  # all game objects + a small Scene interface they depend on
    ├── scene/                   # GameScene, Title/GameOver/LevelStarts, SceneManager
    └── game/                    # ebiten Game wrapper (or leave in root)
```

Dependency flow — one direction only:

```
main → game → scene → entity → engine
                 ↘________↗         ↑
                   assets ──────────┘
highscore ← scene
```

`internal/` signals "app-private" and blocks external imports.

---

## The two structural fixes

### Fix 1 — Break the cycle with a `Scene` interface (declared in `entity`)

Entities stop holding `*GameScene` and hold a narrow interface *declared in the
`entity` package*; `scene.GameScene` satisfies it structurally. Because the
interface lives in `entity`, `scene` imports `entity` (never the reverse).

Keep **all entities in one `entity` package** so intra-entity references
(Shield→Player, Player→Laser/Exhaust/Shield) stay free. The interface is only
for the `entity → scene` boundary. Sketch:

```go
// internal/entity/scene.go
package entity

type Scene interface {
    Space() *resolv.Space
    CheckCollision(obj, against *resolv.Circle) bool

    SpawnLaser(pos engine.Vector, rotation float64)
    SetExhaust(*Exhaust)
    SetShield(*Shield)
    ClearShield()

    // semantic audio — replaces raw *audio.Player reach-in
    PlayThrust(); PauseThrust()
    PlayLaserSound(shot int)
    PlayShieldSound()
}
```

The audio change alone is worth it: today `Player` does
`p.game.thrustPlayer.Rewind(); p.game.thrustPlayer.Play()` in five places.
Collapsing those into `PlayThrust()` / `PlayLaserSound(n)` shrinks the interface
and removes duplication.

### Fix 2 — Expose the fields `GameScene` reaches into

For each field the collision logic touches, either **export the field** (plain
data — meteors, lasers, aliens) or **add a method** (behavior — e.g.
`Player.Kill()` instead of `g.player.isDying = true` scattered across four
handlers). Rough surface: `Meteor{Sprite,Position,Movement,Obj}`,
`Alien{Sprite,Position,Obj,IsIntelligent}`, `Laser{Obj}`,
`AlienLaser{Obj,Position}`, `Player{Obj,Position,Rotation,IsShielded, + lifecycle methods}`.

---

## Phases (each compiles, runs, and is independently committable)

### Phase 0 — Safety net ✅ DONE
- [x] Baseline: `go build ./...` + `go vet ./...` clean
- [x] Unit tests for pure logic (`vector_test.go`, `timer_test.go`)
- [x] Typo fixes: `roataionSpeed*`, `numberOfSmallMeteorsFromLargeMetoer`, `"Wellcome to Hell"`
- [ ] Manual smoke play-test (`go run .`: title → play → die → game-over → restart)
- Deferred to Phase 1: `checkCollision` dead-branch removal (moves to `engine` anyway)

### Phase 1 — Extract leaf packages ⭐ recommended minimum ✅ DONE
- [x] `internal/engine` ← `vector.go`, `timer.go`, `tags.go`, `object-data.go`, `collider.go`
- [x] New `engine/config.go` for `ScreenWidth`/`ScreenHeight` (was in `player.go`)
- [x] Detach `checkCollision` from `*GameScene` → `engine.CheckCollision`, dropped dead `against != nil` branch
- [x] `internal/highscore` ← `helpers.go`, renamed to `Get()` / `Update()`
- [x] `ObjectData.index` → exported `Index` (now cross-package); moved unit tests into `engine`
- [x] Qualified all references; `go build`, `go vet`, `go test`, `gofmt -l` all clean
- Note: local imports currently sort mixed with stdlib (gofmt behavior; `goimports` would regroup — deferrable).

### Phase 2 — Kill global mutable state ✅ DONE
- [x] `highScore`, `originalHighScore` → `GameScene` fields, loaded in `NewGameScene` (was a package `init()` in title-scene.go); `GameOverScene` reads them via `o.game`
- [x] `currentAcceleration`, `shotsFired` (player.go) → `Player` fields
- [x] Fix `NewMeteor(0.25, &GameScene{}, …)` throwaway-scene calls — dropped the `*GameScene` param/field from `Meteor` entirely (it was redundant: the three `m.game` reads all sat inside `GameScene` methods where `m.game == g`). Bonus: `Meteor` now holds no scene reference, making it trivial to move in Phase 3.
- [x] `go build`, `go vet`, `go test`, `gofmt -l` all clean
- Risk: **low-medium** — independent of directories but unblocks a clean split.

### Phase 3 — Extract `entity` package
- [x] **3.1** Move pure entities first: `Star`, `Exhaust`, `AlienLaser`, 3× indicators (no interface needed) — committed separately (2458b62). Along the way: `AlienLaser.Position`/`LaserObj` exported (scene reach-in), `engine.MaxAcceleration` added as the shared thrust/exhaust const, `exhaustSpawnOffset` relocated to its sole user `player.go`.
- [x] **3.2** Move `Meteor` — no scene reference since Phase 2, so it moved as a pure entity (bb6a793). Fields exported: `Obj`/`Sprite`/`Position`/`Movement`; `numberOfSmallMeteorsFromLargeMeteor` relocated to game-scene.go.
- [x] **3.3** Move `Laser` — its `*GameScene` was dead, dropped it; pure move (ab81b0e). Exported `Obj`/`Position`.
- [x] **3.4** Move `Alien` — `*GameScene` used only for `g.player.position` in the ctor, so `NewAlien` now takes the target position; pure move (22d1e1e). Exported `Obj`/`Position`/`Sprite`/`IsIntelligent`.
- [x] **3.5** Extract `Player` + `Shield` behind `entity.Scene` (Fix 1) with semantic audio methods; `GameScene` implements it, asserted via `var _ entity.Scene` (3048bb6). Player exports the ~15 fields the scene reads (Fix 2); Shield holds `*Player` for geometry, space registration moved to `Scene.SetShield`.
- Risk was **medium-high**, concentrated in `Player`. Landed isolated in 3048bb6.
- ⚠️ Not yet smoke-tested — all four splits compile/vet/test clean but the game hasn't been run since Phase 1. Do a manual play-test (title → play → shield → shoot → hyperspace → die → game-over → restart) before relying on it.

### Phase 4 — Extract `scene` and `game` ✅ DONE (7af14ce)
- [x] `internal/scene` ← 4 scenes + `scene-manager.go`. `Input` moved here (scene-manager depends on it → must sit below `game`); added `scene.NewTitleScene()` so `game` builds the first scene without reaching TitleScene's unexported fields.
- [x] `internal/game` ← `game.go`; `main.go` builds `&game.Game{}`.
- [x] `goasteroids` package fully dissolved; `go list -deps` confirms one-directional `main → game → scene → entity → engine`, no back-edges.
- Risk: **low** — mechanical, as predicted.

### Phase 5 — Quality follow-ups (optional)
- [ ] Collapse 3× near-identical `*Indicator` types into one generic `Indicator`
- [ ] Extract the 7× copied "translate/rotate/translate/draw" blit into `engine.DrawSprite(screen, img, pos, rotation)`
- [ ] Consider splitting `game-scene.go` collision handlers into `game-scene-collisions.go`

---

## Recommended stopping point

Best clarity-per-risk: **Phases 0 → 1 → 2, plus the pure-entity slice of Phase 3**,
then reassess. That removes ~10 files from the monolith, centralizes config, and
kills the global-state bug without touching the `Player`/`GameScene` knot.
Phases 3.3–4 (full `entity`/`scene` separation) are worth it only if strict
layering justifies exporting `Player`'s internals — the game works fine with
`Player` and `GameScene` living together.
