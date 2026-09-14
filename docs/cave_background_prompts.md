# Cave Background Generation Prompts

This document contains tailored image generation prompts for each cave type and biome variant in **SubGame**.

These prompts are engineered specifically for **2D side-scrolling atmospheric backgrounds**:
- **Subtle Depth:** Emphasizes deep underwater parallax layers, distant silhouettes, and soft ambient backlighting without harsh contrasts that would distract from foreground tiles or gameplay.
- **Clean Center & Play Area:** Keeps foreground elements and focal clutter absent so player sprites, resources, and creatures pop clearly.
- **Thematic Fidelity:** Directly matches the palette, lighting, fauna, and environmental characteristics defined in `spec.go`, `biome_config.go`, and the individual cave engine scripts.
- **Aspect Ratio:** 16:9 (Native game resolution: `1280x720`).

---

## Quick Reference / Index

1. [Shallow Seabed — Coral Reef (Default)](#1-shallow-seabed--coral-reef)
2. [Shallow Seabed — Kelp Forest](#2-shallow-seabed--kelp-forest)
3. [Shallow Seabed — Thermal Barrens](#3-shallow-seabed--thermal-barrens)
4. [Shallow Seabed — Abyssal Chasm Rim](#4-shallow-seabed--abyssal-chasm-rim)
5. [Deep Organic Trench — Abyssal Grotto](#5-deep-organic-trench--abyssal-grotto)
6. [Wreckage Corridor — Ship 0: Research Tender](#6-wreckage-corridor--ship-0-research-tender)
7. [Wreckage Corridor — Ship 1: Submersible Transport](#7-wreckage-corridor--ship-1-submersible-transport)
8. [Wreckage Corridor — Ship 2: AetherCorp Flagship](#8-wreckage-corridor--ship-2-aethercorp-flagship)
9. [Shock Kelp Cave — Deep Electric Grotto](#9-shock-kelp-cave--deep-electric-grotto)
10. [Thermo Cave — Volcanic Basalt Cavern](#10-thermo-cave--volcanic-basalt-cavern)
11. [The Void — Endless Ocean Abyss](#11-the-void--endless-ocean-abyss)

---

## Universal Negative Prompt & Style Parameters

When using image generators (Midjourney, Stable Diffusion, DALL-E, Imagen), append these parameters/negative constraints:

> **Recommended Settings:**
> - **Aspect Ratio:** `16:9` (`--ar 16:9`)
> - **Style:** Atmospheric 2D side-scrolling video game parallax background, digital matte painting, clean composition, soft atmospheric perspective.
> - **Negative Prompt / Avoid:** `characters, player, submarine, scuba diver, foreground obstacles, solid walls blocking view, UI, HUD, text, watermarks, frames, borders, high contrast foreground clutter, busy patterns, photorealistic human faces`.

---

## 1. Shallow Seabed — Coral Reef

- **Cave Type:** `CaveOrganicShallow` (`ShallowReefBiome`)
- **Depth:** 0m – 30m
- **Theme:** Sunlit turquoise ocean depth, peaceful sandy seabed ridges, distant soft coral arches, and shimmering gentle god rays.
- **Target Filename:** `assets/backgrounds/bg_cave_shallow_reef.png`

### Generation Prompt
```text
Atmospheric 2D side-scrolling game background of a shallow tropical underwater coral reef cavern, distant view. Soft glowing turquoise and cyan water depth, gentle sunbeams filtering down through clear blue sea from the surface above. Distant hazy silhouettes of smooth limestone sea caves, blurred coral outcroppings, and gentle sandy ridges in the far background. Subtle ambient marine particles and tiny bubbles floating in water. Low contrast, soft focus background depth, calm peaceful oceanic mood, clean open center space for gameplay, no characters, no foreground obstruction, panoramic 16:9 composition, digital concept art.
```

---

## 2. Shallow Seabed — Kelp Forest

- **Cave Type:** `CaveOrganicShallow` (`KelpForestBiome`)
- **Depth:** 10m – 40m
- **Theme:** Murky emerald-tinted water, majestic blurred towering kelp stalks fading into emerald fog, filtered diffuse green light shafts.
- **Target Filename:** `assets/backgrounds/bg_cave_kelp_forest.png`

### Generation Prompt
```text
Atmospheric 2D side-scrolling game background of a deep temperate kelp forest underwater cavern, panoramic perspective. Deep emerald green and teal water atmosphere, diffused pale green light filtering through towering silhouettes of giant kelp stalks fading into a soft misty distance. Subtle layered mossy rock shelves and cavernous arches in the far background, faint floating plankton particles and soft micro-bubbles. Subdued ambient contrast, serene mysterious underwater atmosphere, open central space, no foreground obstruction, no characters, 16:9 aspect ratio, atmospheric environmental digital art.
```

---

## 3. Shallow Seabed — Thermal Barrens

- **Cave Type:** `CaveOrganicShallow` (`ThermalBarrensBiome`)
- **Depth:** 15m – 45m
- **Theme:** Murky amber-reddish waters, warm geothermal haze, distant dark volcanic basalt columns, subtle heat distortion.
- **Target Filename:** `assets/backgrounds/bg_cave_thermal_barrens.png`

### Generation Prompt
```text
Atmospheric 2D side-scrolling game background of a geothermal volcanic ocean seabed, distant panoramic perspective. Murky dark amber, dusky ochre, and warm charcoal water tones. In the far distance, silhouettes of dark basalt pillars and volcanic rock ridges fade into hazy mineral water. Faint dull reddish-orange ambient glow rising from deep geothermal fissures far below. Subtle heat shimmer distortion and tiny drifting mineral dust specks. Low contrast, atmospheric and moody, completely open central space for gameplay, no foreground obstacles, no characters, 16:9 widescreen, digital matte painting.
```

---

## 4. Shallow Seabed — Abyssal Chasm Rim

- **Cave Type:** `CaveOrganicShallow` (`AbyssalBlueBiome`)
- **Depth:** 30m – 60m
- **Theme:** Deep indigo-slate water, dramatic drop-off chasm rim, distant dark rock shelves with faint crystalline bioluminescence.
- **Target Filename:** `assets/backgrounds/bg_cave_chasm_rim.png`

### Generation Prompt
```text
Atmospheric 2D side-scrolling game background of a deep oceanic shelf and subterranean chasm rim, panoramic perspective. Rich midnight blue, slate, and deep indigo water fading into a profound dark trench abyss below. In the far background, silhouetted terraced rock shelves and monolithic undersea cliff drop-offs. Extremely subtle faint cyan-teal crystalline bioluminescent glimmer along distant rock fractures. Soft ambient deep-sea particulate, high atmospheric depth, low foreground contrast, clean center composition, no characters, no HUD, 16:9 aspect ratio, cinematic video game environment art.
```

---

## 5. Deep Organic Trench — Abyssal Grotto

- **Cave Type:** `CaveOrganicTrench`
- **Depth:** 40m – 120m+ (Luminous Pneumatophore Grotto down to Benthic Brine-Falls)
- **Theme:** Monolithic abyss, crushing deep ocean pressure, dark indigo-black water, ethereal bioluminescent azure/cyan veins along distant trench walls.
- **Target Filename:** `assets/backgrounds/bg_cave_organic_trench.png`

### Generation Prompt
```text
Atmospheric 2D side-scrolling game background of a massive deep-sea abyssal trench grotto, panoramic view. Crushing deep-sea environment with midnight indigo and obsidian water. In the far distance, colossal sheer subterranean rock walls and chasms descend into an endless dark oceanic abyss. Ethereal, faint bioluminescent cyan and electric azure veins softly glowing along distant mineral strata. Sparse floating marine snow and microscopic glowing plankton drifting slowly. Vast scale, quiet intimidating depth, subdued ambient lighting, open center, no characters, no foreground obstruction, 16:9 cinematic environmental art.
```

---

## 6. Wreckage Corridor — Ship 0: Research Tender

- **Cave Type:** `CaveWreckage` (`ShipIndex = 0`)
- **Depth:** 30m – 40m
- **Theme:** Clinical sunken sci-fi science ship, cold cyan-grey steel hull bulkheads, submerged scientific consoles, eerie quiet.
- **Target Filename:** `assets/backgrounds/bg_cave_wreckage_tender.png`

### Generation Prompt
```text
Atmospheric 2D side-scrolling game background inside the flooded hull of a sunken futuristic scientific research starship, distant interior perspective. Submerged dark cyan-steel bulkheads, modular laboratory corridor ribs, and flooded structural framework fading into murky blue-grey water. Distant emergency terminal monitors casting a very soft, faint teal glow through the dark water. Subdued underwater haze with slow-drifting metallic micro-particles. Low contrast background layer, spacious and desolate, clean open central path, no characters, no HUD, 16:9 widescreen, digital concept art.
```

---

## 7. Wreckage Corridor — Ship 1: Submersible Transport

- **Cave Type:** `CaveWreckage` (`ShipIndex = 1`)
- **Depth:** 40m – 60m+
- **Theme:** Heavy industrial transport vessel, rusted cast iron plating, submerged cargo gantry framework, dim amber emergency beacon glow.
- **Target Filename:** `assets/backgrounds/bg_cave_wreckage_transport.png`

### Generation Prompt
```text
Atmospheric 2D side-scrolling game background inside a flooded industrial cargo submarine wreckage, distant perspective. Submerged rusted iron girders, heavy structural frames, and deep industrial maintenance chambers fading into dark murky brownish-grey water. In the far background, a faint dying amber emergency bulkhead indicator provides a subtle warm glow in the darkness. Distant silt and rust particles suspended in still water. Claustrophobic industrial mood, low contrast background layer, open central space, no characters, no UI, 16:9 aspect ratio, atmospheric matte painting.
```

---

## 8. Wreckage Corridor — Ship 2: AetherCorp Flagship

- **Cave Type:** `CaveWreckage` (`ShipIndex = 2`)
- **Depth:** 60m – 100m+
- **Theme:** Imposing obsidian dreadnought armor, sunken command decks, deep crimson emergency light strips, ominous high-tech vault aura.
- **Target Filename:** `assets/backgrounds/bg_cave_wreckage_flagship.png`

### Generation Prompt
```text
Atmospheric 2D side-scrolling game background deep inside the flooded wreckage of an imposing high-tech military flagship starship, distant perspective. Submerged sleek obsidian-black armor plating, heavy reinforced vault bulkheads, and deep angular corridors fading into near-pitch black abyssal water. Very faint, ominous crimson-red emergency light conduits glowing dimly in the distant background. Subtle deep-water atmospheric haze, sleek sci-fi shipwreck architecture, dark and intimidating, low contrast background, open center, no characters, 16:9 widescreen digital art.
```

---

## 9. Shock Kelp Cave — Deep Electric Grotto

- **Cave Type:** `CaveShockKelp`
- **Depth:** 60m – 80m
- **Theme:** Dark violet-purple cavern depths, bioluminescent electric cyan tendrils, floating charged parallax motes, alien reef aura.
- **Target Filename:** `assets/backgrounds/bg_cave_shock_kelp.png`

### Generation Prompt
```text
Atmospheric 2D side-scrolling game background of a deep subterranean bioluminescent shock-kelp grotto, panoramic perspective. Dark purple-grey and deep violet ambient water. In the far background, distant shadowy rock spires draped with soft glowing electric cyan and neon-purple bioluminescent vines. Faint glowing electric spores and ionized particulate gently drifting in the deep water. Ethereal, alien, mysterious atmosphere with very subtle low-intensity luminescence, open center for player movement, no characters, no foreground obstruction, 16:9 widescreen, video game environment concept art.
```

---

## 10. Thermo Cave — Volcanic Basalt Cavern

- **Cave Type:** `CaveThermo`
- **Depth:** 40m – 50m
- **Theme:** Dark volcanic basalt cavern, molten magma veins glowing dull crimson in the distance, geothermal smoker silhouettes, rising orange embers.
- **Target Filename:** `assets/backgrounds/bg_cave_thermo.png`

### Generation Prompt
```text
Atmospheric 2D side-scrolling game background of a deep subterranean volcanic sea cave, panoramic perspective. Dark charcoal basalt rock walls and jagged cavern silhouettes fading into murky dark red-tinted water. In the distant background, narrow fissures in the rock glow with deep molten crimson and warm incandescent orange. Silhouettes of distant hydrothermal smoker chimneys, with tiny faint glowing orange embers floating upward through the thermal currents. Oppressive geothermal mood, soft background lighting, open central area, no characters, no UI, 16:9 aspect ratio, digital painting.
```

---

## 11. The Void — Endless Ocean Abyss

- **Cave Type:** `CaveVoid`
- **Depth:** 150m+ (Crushing boundless abyss)
- **Theme:** Terrifying oceanic void, nearly pure black gradient, faint distant deep indigo mist, floating abyssal dust.
- **Target Filename:** `assets/backgrounds/bg_cave_void.png`

### Generation Prompt
```text
Atmospheric 2D side-scrolling game background of an infinite oceanic abyss, panoramic view. Extreme deep sea minimalism: pitch-black crushing water with only the faintest, subtle gradient of dark midnight-indigo fading into pure blackness towards the bottom. Extremely sparse, faint microscopic bioluminescent marine dust particles drifting in the vast empty ocean expanse. Boundless, silent, terrifying depth and scale, completely empty center, no solid ground, no characters, no UI, 16:9 widescreen, cinematic atmospheric digital art.
```

---

## Implementation Guide (Ebitengine)

Once the images are generated and saved to `assets/backgrounds/`, they can be rendered with parallax scrolling in `DrawBackground`:

```go
// Example integration pattern in DrawBackground
func (c *ShallowSeabedCave) DrawBackground(screen *ebiten.Image, camY float64, maxDepth float64, lightMult float64) {
    if c.bgImage != nil {
        op := &ebiten.DrawImageOptions{}
        // Parallax vertical scrolling at 20% camera speed:
        parallaxY := -math.Mod(camY * 0.20, float64(c.bgImage.Bounds().Dy()))
        op.GeoM.Translate(0, parallaxY)
        // Tint with current ambient light multiplier:
        op.ColorScale.Scale(float32(lightMult), float32(lightMult), float32(lightMult), 1.0)
        screen.DrawImage(c.bgImage, op)
        return
    }
    // ... fallback gradient ...
}
```
