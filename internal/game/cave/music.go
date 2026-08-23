package cave

// MusicTrack returns the looping music path for a cave type.
func MusicTrack(t CaveType) string {
	switch t {
	case CaveThermo:
		return "music/cave_volcanic.mp3"
	case CaveShockKelp:
		return "music/cave_kelp.mp3"
	case CaveWreckage:
		return "music/cave_wreckage.mp3"
	case CaveVoid, CaveOrganicTrench:
		return "music/cave_abyssal.mp3"
	default:
		return "music/cave_shallow.mp3"
	}
}
