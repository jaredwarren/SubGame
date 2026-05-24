# **2D Subnautica-Inspired Game Design Blueprint**

## **1\. The Dual-Perspective Gameplay Loop**

To make the two perspectives feel seamless, they should serve distinct mechanical purposes:

* **The Overworld (Top-Down):** This is the surface. You pilot a floating life pod or a surface boat. It’s safer, navigation is easy, and you use it to travel between major biomes (e.g., Safe Shallows, Coral Reefs, Volcanic Trenches). You scan the surface to find massive "sinkholes" or "trench lines" that act as transition points.  
* **The Caves (Side-View):** Dropping into a trench switches the camera to a side-scroller. This is where the real game happens. Physics change to 2D swimming (buoyancy, inertia). You manage oxygen, dodge predators, navigate tight tunnels, and mine resources embedded in the cave walls.

## **2\. Adapting Subnautica Elements to 2D**

### **The Terror of the Dark (Dynamic Lighting)**

In 3D, fear comes from what's lurking behind you. In a 2D side-scroller, fear comes from **limited sightlines**.

* Implement a tight **Line-of-Sight light cone** attached to the player’s flashlight.  
* If you face right, everything to your left falls into pitch blackness.  
* Bioluminescent plants and glowing eyes of predators can flash briefly in the dark just outside your beam.

### **Oxygen & Buoyancy**

Instead of just a timer, oxygen and physics can intertwine.

* **Buoyancy:** Early game, you naturally drift upward toward the cave ceiling if you stop swimming.  
* **Pressure:** Going past certain depths without a reinforced suit drains oxygen faster or cracks your vehicle’s hull.

## **3\. Procedural Generation Strategy**

Generating a hybrid world can be achieved using a layered grid approach:

\[ Overworld: 2D Perlin Noise Map \]  
       |         |         |  
   \[Sinkhole\] \[Trench\]  \[Sinkhole\]  
       |         |         |  
\[ Cave Layer 1: Cellular Automata (Shallow Caves) \]  
       |  
\[ Cave Layer 2: Drunkard's Walk (Deep Abyssal Crevices) \]  
    

1. **The Surface Layer:** Use standard 2D Perlin or Simplex noise to generate an island/ocean map. "Land" can be impassable reefs, while ultra-dark pixels dictate where the deep trenches open up.  
2. **The Cave Transition:** When the player overlaps with a trench tile on the surface, you load a specific side-view cave map linked to those coordinates.  
3. **The Cave Structures:**  
   * For open, organic-feeling caves, use **Cellular Automata** (smooth, bubble-like underground pockets).  
   * For deep, winding, claustrophobic tunnels that drill straight down, use a **Drunkard's Walk (Random Walk)** algorithm that is heavily weighted to dig downward, creating vertical labyrinth structures.

## **4\. Upgrade & Vehicle Blueprint**

Progression in this style of game is all about **"Gatekeeping by Depth."** You can't go deeper until you upgrade your gear to survive the pressure and the darkness.

### **Personal Gear Upgrades**

* **O2 Tanks:** Standard → High Capacity → Ultra High Capacity. (Adds more tiles of exploration time).  
* **Fins & Propulsion:** Increases swim speed and counteracts heavy currents in narrow cave bottlenecks.  
* **The Scanner:** Reveals resource nodes behind breakable tiles or maps out a small radius of the cave through the fog of war.

### **The Vehicle Hierarchy**

Because you have two perspectives, vehicles can be tailored to the view, providing entirely different styles of gameplay.

| Vehicle | Perspective | Role & Function | Unique Upgrade Module   |
| :---- | :---- | :---- | :---- |
| **The Skiff** | Top-Down | A surface boat used to haul massive amounts of cargo between distant trench openings. Acts as a mobile surface base. | **Solar Charging:** Recharges batteries when parked in sunny overworld biomes. |
| **The Scout Sub** | Side-View | A small, highly agile mini-sub. Can fit through tight 2-tile-wide cave passages where the player would otherwise risk drowning. | **Sonar Ping:** Emits a shockwave that briefly lights up the entire 2D screen, revealing hidden walls and predators. |
| **The Heavy Mech** | Side-View | A heavy suit that ignores buoyancy and sinks straight to the bottom. Immune to minor predator bites. | **Drill Arm:** Required to break open massive "Abyssal Ore Blocks" that are otherwise indestructible. |

## **5\. Base Building System: The Modular Management Menu**

Base building in this game is focused entirely on utility rather than architectural simulation. To maintain the fast pacing and tension of the game loop, bases exist as abstract **Menu Hubs** rather than physical structures built tile-by-tile.

### **Core Concept: The Anchor Terminal**

Your base is represented in the world by a central hub—such as your starting Life Pod, your surface boat (The Skiff), or deployable "Anchor Modules." When you approach this hub and interact with it, you enter the Management Screen. There is no grid-snapping or physical wall construction; the focus is on installing modules and managing resources.

### **How the Management Menu Works**

* **The Hub Schematic:** Upon interaction, you see a clean, multi-tab UI displaying a blueprint schematic of your base/vehicle. The schematic contains a set number of "Module Slots."  
* **Direct Upgrading:** You spend gathered resources (e.g., Titanium, Copper, Quartz) directly within the menu to purchase and install functional modules, such as an "Infirmary Module" or an "Expanded Cargo Hold."  
* **Functional Tabs:** Once a module is installed, its corresponding tab becomes active for use:  
  * *Fabricator Tab:* The central crafting hub for processing raw ores, building tools, upgrading O2 tanks, and creating vehicle parts.  
  * *Medical Tab:* Utilizes base energy to slowly generate health packs or cure negative status effects (like deep-sea venom).  
  * *Storage Tab:* A massive, scrollable grid inventory where you can hoard resources gathered from your deep-sea dives.

### **Advantages of the Menu System**

* **Streamlined Development:** Removes the need to program complex 2D block collision, interior pathfinding, or structural integrity logic.  
* **Focus on the Core Loop:** The player is encouraged to quickly drop off loot, craft upgrades, and get right back into the danger of the caves.  
* **Sense of Safety:** The UI screen itself becomes a psychological safe haven, giving the player a breather from the tense, line-of-sight cave exploration.

## **6\. Deep-Sea Biomes and Predators**

These alien but biologically realistic ecosystems rely on convergent evolution, with life adapting to high pressure, extreme darkness, and fluid dynamics.

### **Biome 1: The Luminous Pneumatophore Grotto (Mid-Depth)**

**The Concept:** A transition zone where the caves are choked with massive, balloon-like flora that use trapped gases for buoyancy and bioluminescence to attract symbiotic microbes.  
**Visual Palette:** Deep teals, bioluminescent cyan, and soft pinks. The background is filled with overlapping, translucent bubbles.

* **The Flora \- "Shatter-Bulbs":** Large, fragile, glass-like bladders clinging to the cave walls. They emit a soft glow.  
  * *2D Mechanic:* If you bump into them, they pop, releasing a burst of oxygen (which you can catch) but also creating a loud "snap" that alerts nearby predators.  
* **The Predator \- The False-Bulb Snare (Ambush)**  
  * *Biology:* A cephalopod-like creature with a hydrostatic skeleton. It hangs upside down from the ceiling, its mantle perfectly mimicking a glowing Shatter-Bulb.  
  * *Behavior:* It hunts based on **light aversion**. When your flashlight beam is directly on it, it freezes, perfectly camouflaged. The moment you turn your flashlight *away* (leaving it in the dark edge of your screen), it lunges rapidly across the 2D plane to grapple you. You have to "keep your eyes on it" while swimming backward to escape.

### **Biome 2: The Silicate Smoker Trenches (Deep)**

**The Concept:** An incredibly hostile, high-heat biome driven by chemosynthesis. Instead of carbon-based plants, the flora here are slow-growing silicate (glass) structures that feed on superheated mineral vents.  
**Visual Palette:** Vantablack, harsh jagged grays, and blinding thermal oranges/yellows. Water here shimmers with heat distortion.

* **The Flora \- "Brimstone Siphons":** Razor-sharp, hollow crystalline tubes jutting horizontally from the 2D walls, venting scalding water in rhythmic pulses.  
  * *2D Mechanic:* They act as environmental hazards. You must time your swimming to bypass their heat-jets, but their thick mineral crusts hold highly valuable mid-game crafting ores.  
* **The Predator \- The Thermocline Rammer (Pursuit)**  
  * *Biology:* A heavily armored, eyeless crustacean/fish hybrid. It has a massive, shovel-like head made of biological iron-sulfide armor (similar to Earth's scaly-foot gastropod).  
  * *Behavior:* It is completely blind and ignores your flashlight. Instead, it hunts via **heat and vibration**. If you use your vehicle's thrusters or standard fast-swimming in its territory, it detects you. It attacks by charging in a straight, high-speed horizontal or vertical line (like a rook in chess). You must use 2D verticality to dodge its charge and let it crash into the cave walls, temporarily stunning it.

### **Biome 3: The Benthic Brine-Falls (Abyssal)**

**The Concept:** At the bottom of the world, heavy, super-saline water pools into underwater "lakes" and "waterfalls" within the ocean itself. It is eerily quiet, dead, and fleshy. The walls are covered in pale, filtering organic mats.  
**Visual Palette:** Monochrome. Ash-white, pale yellows, and complete, crushing darkness.

* **The Flora \- "Pallid Nerve-Mats":** Pale, hair-like root systems that cover the cave floors and ceilings. They feed on falling detritus ("marine snow").  
  * *2D Mechanic:* Swimming too close to the mats slows you down as the tendrils weakly cling to your suit.  
* **The Predator \- The Electro-Weaver (Apex / Stalker)**  
  * *Biology:* A massive, serpentine creature that resembles a ribbon eel mixed with a deep-sea siphonophore. It has no face, just a glowing, fractal nervous system visible through perfectly transparent skin.  
  * *Behavior:* The Weaver hunts via **electroreception**. It does not attack aggressively. Instead, it stalks. If you use your vehicle's "Sonar Ping" or keep your flashlight turned on for too long, the Weaver detects the electrical output.  
  * *The Scare:* You won't hear it coming. Instead, your screen's UI will start to glitch, and your light beam will flicker. You will slowly see its massive, transparent, glowing body sliding silently through the background layer of the 2D cave, before it loops into the foreground to strike. To survive, you must cut your vehicle's power, turn off your lights, and drift in the pitch black until it loses interest.

## **7\. Implementation Considerations for Go & Ebitengine**

* **State Management:** Utilize a state machine to cleanly transition between the Overworld state, the Cave state, and the newly added Base Menu state. Since Ebitengine's Update() and Draw() loop runs continuously, segregating these states ensures you aren't calculating 2D water physics while the player is managing inventory.  
* **Lighting and Shaders:** For the "Line-of-Sight light cone" in the 2D side-view, use Kage (Ebitengine's shading language). Pass the player's coordinates, light radius, and direction as uniforms to a custom fragment shader to darken pixels outside the light cone.  
* **Procedural Generation:** Go is highly performant, making real-time map generation using 2D arrays and noise libraries (like Perlin noise modules) very efficient. Pre-generate chunks of the Cellular Automata or Drunkard's Walk caves in a separate Goroutine as the player approaches a transition point.  
* **UI Architecture:** Use an immediate mode GUI library compatible with Ebitengine (like Egui or a custom-built widget system) to handle the complex rendering and interaction logic of the Base Management Menu tabs.

Other ideas

* Depth sensor upgrade, fog for deeper areas   
* Start with a life pod, can't move. Build better faster vehicles allow access to deeper caves  
* 2 stamina or energy bars, Max and current? Max slowly decreased unless rest, eat, etc. “current” behaves like dark souls, regenerates.  
* Health bar  
* Air bar.   
* 