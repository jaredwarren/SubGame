package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// LightConeShaderCode contains the Kage fragment shader string compiled at runtime.
const LightConeShaderCode = `
package main

var LightSource vec2
var FlashlightDir vec2
var LightRadius float
var ConeHalfAngle float
var SonarSource vec2
var SonarRadius float
var PersonalRadius float
var AmbientColor vec4
var EntranceLight vec2
var EntranceActive float

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
	pixelPos := position.xy
	toPixel := pixelPos - LightSource
	dist := length(toPixel)

	// Personal ambient circle surrounding the player
	personalIntensity := 0.0
	if dist < PersonalRadius {
		personalIntensity = 1.0 - (dist / PersonalRadius)
		personalIntensity = personalIntensity * personalIntensity * (3.0 - 2.0*personalIntensity)
	}

	// Directional flashlight cone beam
	coneIntensity := 0.0
	if dist < LightRadius && dist > 0.0 {
		dirToPixel := toPixel / dist
		dotVal := dot(dirToPixel, FlashlightDir)
		minDot := cos(ConeHalfAngle)

		if dotVal >= minDot {
			radialFade := 1.0 - (dist / LightRadius)
			radialFade = radialFade * radialFade * (3.0 - 2.0*radialFade)

			angularFade := (dotVal - minDot) / (1.0 - minDot)
			angularFade = clamp(angularFade * 6.0, 0.0, 1.0)

			coneIntensity = radialFade * angularFade
		}
	}

	// Sonar wave illumination
	sonarIntensity := 0.0
	if SonarRadius > 0.0 {
		toSonar := pixelPos - SonarSource
		distSonar := length(toSonar)
		if distSonar < SonarRadius {
			ageFactor := 1.0 - (SonarRadius / 600.0)
			if ageFactor > 0.0 {
				if distSonar > SonarRadius - 30.0 {
					sonarIntensity = 0.85 * ageFactor
				} else {
					sonarIntensity = 0.3 * (distSonar / SonarRadius) * ageFactor
				}
			}
		}
	}

	// Entrance light cone beam (pointing straight down)
	entranceIntensity := 0.0
	if EntranceActive > 0.0 {
		toEntrance := pixelPos - EntranceLight
		distEntrance := length(toEntrance)
		if distEntrance < 700.0 && distEntrance > 0.0 {
			dirToEntrance := toEntrance / distEntrance
			dotVal := dot(dirToEntrance, vec2(0.0, 1.0))
			minDot := cos(0.65) // wide cone angle (~74 degrees total span)

			if dotVal >= minDot {
				radialFade := 1.0 - (distEntrance / 700.0)
				radialFade = radialFade * radialFade * (3.0 - 2.0*radialFade)

				angularFade := (dotVal - minDot) / (1.0 - minDot)
				angularFade = clamp(angularFade * 4.0, 0.0, 1.0)

				entranceIntensity = radialFade * angularFade * 1.0
			}
		}
	}

	// Blend illumination channels
	totalLight := max(personalIntensity, max(coneIntensity, max(sonarIntensity, entranceIntensity)))
	totalLight = clamp(totalLight, 0.0, 1.0)

	// Dark overlay mask: fully lit areas remain transparent, dark areas become ambient color
	return AmbientColor * (1.0 - totalLight)
}
`

// LightShader is the compiled flashlight shader instance.
var LightShader *ebiten.Shader

func init() {
	var err error
	LightShader, err = ebiten.NewShader([]byte(LightConeShaderCode))
	if err != nil {
		panic("failed to compile Kage light cone shader: " + err.Error())
	}
}
