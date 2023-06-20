package center

type Flag int

type CenterOptions struct {
	Flag Flag
}

type CenterOption func(opts *CenterOptions)
