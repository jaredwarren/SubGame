package shader

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// LightConeShaderCode contains the Kage fragment shader string compiled at runtime.
const LightConeShaderCode = `
//kage:unit pixels
package main

var LightSource vec2
var FlashlightDir vec2
var LightRadius float
var ConeHalfAngle float
var SonarSource vec2
var SonarRadius float
var SonarBright float
var SonarFadeLimit float
var PersonalRadius float
var AmbientColor vec4
var EntranceLight vec2
var EntranceActive float
var LavaPositions [8]vec2
var LavaCount float
var BiolumPositions [16]vec4
var BiolumColors [16]vec4
var BiolumCount float

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	pixelPos := position.xy - imageDstOrigin()
	toPixel := pixelPos - LightSource
	dist := length(toPixel)

	// Personal ambient circle surrounding the player
	personalIntensity := 0.0
	if dist < PersonalRadius {
		personalIntensity = 1.0 - (dist / max(PersonalRadius, 0.001))
		personalIntensity = personalIntensity * personalIntensity * (3.0 - 2.0*personalIntensity)
	}

	// Directional flashlight cone beam
	coneIntensity := 0.0
	if dist < LightRadius && dist > 0.0 {
		dirToPixel := toPixel / max(dist, 0.001)
		dotVal := dot(dirToPixel, FlashlightDir)
		minDot := cos(ConeHalfAngle)
		outerDot := cos(ConeHalfAngle + 0.12) // 0.12 rad soft spill edge (~7 degrees)

		if dotVal >= outerDot {
			radialFade := 1.0 - (dist / max(LightRadius, 0.001))
			radialFade = radialFade * radialFade * (3.0 - 2.0*radialFade)

			// Smooth transition from outerDot (0.0) to minDot (1.0)
			angularFade := (dotVal - outerDot) / max(minDot - outerDot, 0.001)
			angularFade = clamp(angularFade, 0.0, 1.0)
			angularFade = angularFade * angularFade * (3.0 - 2.0*angularFade) // smoothstep

			coneIntensity = radialFade * angularFade
		}
	}

	// Sonar wave illumination
	sonarIntensity := 0.0
	if SonarRadius > 0.0 {
		toSonar := pixelPos - SonarSource
		distSonar := length(toSonar)
		if distSonar < SonarRadius {
			limit := SonarFadeLimit
			if limit <= 0.0 {
				limit = 800.0
			}
			ageFactor := 1.0 - (SonarRadius / limit)
			if ageFactor > 0.0 {
				if distSonar > SonarRadius - 30.0 {
					sonarIntensity = 1.0 * ageFactor
				} else {
					sonarIntensity = 0.72 * (distSonar / max(SonarRadius, 0.001)) * ageFactor
				}
				multiplier := SonarBright
				if multiplier <= 0.0 {
					multiplier = 1.0
				}
				sonarIntensity = sonarIntensity * multiplier
			}
		}
	}

	// Entrance light cone beam (pointing straight down)
	entranceIntensity := 0.0
	if EntranceActive > 0.0 {
		toEntrance := pixelPos - EntranceLight
		distEntrance := length(toEntrance)
		if distEntrance < 700.0 && distEntrance > 0.0 {
			dirToEntrance := toEntrance / max(distEntrance, 0.001)
			dotVal := dot(dirToEntrance, vec2(0.0, 1.0))
			minDot := cos(0.65) // wide cone angle (~74 degrees total span)

			if dotVal >= minDot {
				radialFade := 1.0 - (distEntrance / 700.0)
				radialFade = radialFade * radialFade * (3.0 - 2.0*radialFade)

				angularFade := (dotVal - minDot) / (1.0 - minDot)
				angularFade = clamp(angularFade * 4.0, 0.0, 1.0)

				entranceIntensity = radialFade * angularFade
			}
		}
	}

	// Lava lights
	lavaIntensity := 0.0
	lCount := int(LavaCount)
	for i := 0; i < 8; i++ {
		if i >= lCount {
			break
		}
		toLava := pixelPos - LavaPositions[i]
		distLava := length(toLava)
		if distLava < 180.0 {
			// Radial fade for lava glow
			fade := 1.0 - (distLava / 180.0)
			fade = fade * fade * (3.0 - 2.0*fade)
			lavaIntensity = max(lavaIntensity, fade * 0.95)
		}
	}

	// Bioluminescent creature point lights
	biolumIntensity := 0.0
	biolumTint := vec3(0.0)
	bCount := int(BiolumCount)
	for i := 0; i < 16; i++ {
		if i >= bCount {
			break
		}
		toBiolum := pixelPos - BiolumPositions[i].xy
		distBiolum := length(toBiolum)
		radBiolum := BiolumPositions[i].z
		if distBiolum < radBiolum && radBiolum > 0.0 {
			fade := 1.0 - (distBiolum / radBiolum)
			fade = fade * fade * (3.0 - 2.0*fade)
			intensity := fade * BiolumPositions[i].w
			if intensity > biolumIntensity {
				biolumIntensity = intensity
			}
			biolumTint += BiolumColors[i].rgb * intensity
		}
	}

	// Blend illumination channels
	totalLight := max(personalIntensity, max(coneIntensity, max(sonarIntensity, max(entranceIntensity, max(lavaIntensity, biolumIntensity)))))
	totalLight = clamp(totalLight, 0.0, 1.0)

	// Colored atmospheric tint on the fringes of bioluminescent point lights
	maskColor := AmbientColor.rgb
	if biolumIntensity > 0.001 {
		normTint := biolumTint / max(biolumIntensity, 0.001)
		tintFactor := clamp(biolumIntensity * 0.35, 0.0, 0.50)
		maskColor = mix(maskColor, normTint, tintFactor)
	}

	// Dark overlay mask: fully lit areas remain transparent, dark areas become ambient color
	return vec4(maskColor, AmbientColor.a * (1.0 - totalLight))
}
`

// LightShader is the compiled flashlight shader instance.
var LightShader *ebiten.Shader

// WaterDisplacementShader is the compiled screen water ripple shader.
var WaterDisplacementShader *ebiten.Shader

// EdgeBlurShader is the compiled soft screen-edge defocus shader (cave only).
var EdgeBlurShader *ebiten.Shader

// WaterDisplacementShaderCode contains the Kage fragment shader for screen water shimmer and localized volcanic heat waves.
const WaterDisplacementShaderCode = `
//kage:unit pixels
package main

var Time float
var VentPositions [8]vec2
var VentCount float
var SurfaceY float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	offsetX := 0.0
	offsetY := 0.0

	// Only apply water shimmer below the water surface Y coordinate
	if dstPos.y >= SurfaceY {
		// Base water shimmer displacement (reduced from 2.0 to 0.8)
		offsetX = sin(srcPos.y * 0.04 + Time * 0.04) * 0.8
		offsetY = cos(srcPos.x * 0.04 + Time * 0.04) * 0.8

		// Add a secondary wave for more fluid variety (reduced from 0.8 to 0.3)
		offsetX += sin(srcPos.y * 0.09 - Time * 0.06) * 0.3
		offsetY += cos(srcPos.x * 0.09 - Time * 0.06) * 0.3
	}

	// Add high-frequency heat wave distortion from active Brimstone Siphons
	vCount := int(VentCount)
	for i := 0; i < 8; i++ {
		if i >= vCount {
			break
		}
		ventPos := VentPositions[i]
		dist := distance(dstPos.xy, ventPos)

		// Volcanic vents produce localized turbulent heat wave shimmers within 180px
		if dist < 180.0 {
			// Fades out with distance
			factor := 1.0 - (dist / 180.0)
			factor = factor * factor // square for nicer falloff
			
			// Heat distortion is much faster and has higher spatial frequency
			heatOffset := sin(dstPos.y * 0.15 - Time * 0.25) * 5.0 * factor
			offsetX += heatOffset
			offsetY += heatOffset
		}
	}

	distortedCoords := srcPos + vec2(offsetX, offsetY)

	// Clamp to destination image bounds to avoid rendering artifacts at borders
	origin := imageSrc0Origin()
	size := imageSrc0Size()
	distortedCoords = clamp(distortedCoords, origin + vec2(0.5), origin + size - vec2(0.5))

	return imageSrc0At(distortedCoords)
}
`

// EdgeBlurShaderCode soft-blurs pixels near the frame so the cave view reads
// slightly unfocused at the edges (visor / pressure feel). Center stays sharp.
// This is real sample blur of the scene, not a dark geometric overlay.
const EdgeBlurShaderCode = `
//kage:unit pixels
package main

var FalloffPx float // how far inward the soft edge reaches
var MaxBlurPx float // peak sample radius at the absolute border
var Strength float  // 0..1 overall mix (depth-driven)

func sampleClamped(uv vec2) vec4 {
	origin := imageSrc0Origin()
	size := imageSrc0Size()
	uv = clamp(uv, origin+vec2(0.5), origin+size-vec2(0.5))
	return imageSrc0At(uv)
}

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	origin := imageSrc0Origin()
	size := imageSrc0Size()

	// Distance (px) to nearest screen edge.
	p := dstPos.xy - origin
	edgeDist := min(min(p.x, p.y), min(size.x-p.x, size.y-p.y))

	center := imageSrc0At(srcPos)

	falloff := FalloffPx
	if falloff < 1.0 {
		falloff = 1.0
	}
	if edgeDist >= falloff || Strength <= 0.001 {
		return center
	}

	// 0 deep inside, 1 at the border — smooth so no hard ring/box contour.
	t := 1.0 - edgeDist/falloff
	t = t * t * (3.0 - 2.0*t)
	t = t * Strength

	r := MaxBlurPx * t
	if r < 0.25 {
		return center
	}

	// Dual-ring 17-tap blur: inner ring at r/2 fills the disc; outer at r
	// pushes the soft edge out. Stronger than a single 9-tap ring.
	sum := center
	rHalf := r * 0.5
	sum += sampleClamped(srcPos + vec2(rHalf, 0.0))
	sum += sampleClamped(srcPos + vec2(-rHalf, 0.0))
	sum += sampleClamped(srcPos + vec2(0.0, rHalf))
	sum += sampleClamped(srcPos + vec2(0.0, -rHalf))
	sum += sampleClamped(srcPos + vec2(rHalf, rHalf)*0.707)
	sum += sampleClamped(srcPos + vec2(rHalf, -rHalf)*0.707)
	sum += sampleClamped(srcPos + vec2(-rHalf, rHalf)*0.707)
	sum += sampleClamped(srcPos + vec2(-rHalf, -rHalf)*0.707)
	sum += sampleClamped(srcPos + vec2(r, 0.0))
	sum += sampleClamped(srcPos + vec2(-r, 0.0))
	sum += sampleClamped(srcPos + vec2(0.0, r))
	sum += sampleClamped(srcPos + vec2(0.0, -r))
	sum += sampleClamped(srcPos + vec2(r, r)*0.707)
	sum += sampleClamped(srcPos + vec2(r, -r)*0.707)
	sum += sampleClamped(srcPos + vec2(-r, r)*0.707)
	sum += sampleClamped(srcPos + vec2(-r, -r)*0.707)
	blurred := sum / 17.0

	return mix(center, blurred, t)
}
`

func init() {
	var err error
	LightShader, err = ebiten.NewShader([]byte(LightConeShaderCode))
	if err != nil {
		panic("failed to compile Kage light cone shader: " + err.Error())
	}

	WaterDisplacementShader, err = ebiten.NewShader([]byte(WaterDisplacementShaderCode))
	if err != nil {
		panic("failed to compile Kage water displacement shader: " + err.Error())
	}

	EdgeBlurShader, err = ebiten.NewShader([]byte(EdgeBlurShaderCode))
	if err != nil {
		panic("failed to compile Kage edge blur shader: " + err.Error())
	}
}
