package rotate

import "time"

type Option func(*Rotate)

// WithRotationDuration set time interval, Preferably divisible by 24 Hour.
func WithRotationDuration(RotationDuration time.Duration) Option {
	return func(r *Rotate) {
		r.rotateDur = RotationDuration
	}
}

// WithTimeZone set time zone
func WithTimeZone(loc *time.Location) Option {
	return func(r *Rotate) {
		r.loc = loc
	}
}

// WithMaxAge max survival time
func WithMaxAge(MaxAge time.Duration) Option {
	return func(r *Rotate) {
		r.maxAge = MaxAge
	}
}

// WithMaxSize max file size
func WithMaxSize(MaxSize int64) Option {
	return func(r *Rotate) {
		r.maxSize = MaxSize
	}
}

// WithExpiredHandler expired file handler
func WithExpiredHandler(handler ExpFunc) Option {
	return func(r *Rotate) {
		r.expiredFunc = handler
	}
}

// WithDeleteEmptyFile whether to delete empty files
func WithDeleteEmptyFile(flag bool) Option {
	return func(r *Rotate) {
		r.deleteEmptyFileFlag = flag
	}
}

// WithDeleteEmptyDir whether to delete empty folders
func WithDeleteEmptyDir(flag bool) Option {
	return func(r *Rotate) {
		r.deleteEmptyDirFlag = flag
	}
}
