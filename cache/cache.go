package cache

type CacheOptions struct {
	Path string
}

type Cache interface {
	ArtifactCache
	PlanCache
	PrimordialCache
}

type ArtifactCache interface {
	// Return `true` if *all* of the output Artifacts of the Rule are stored in the Cache, otherwise return `false`.
	HasArtifacts(*Rule) bool
	// Store the output Artifacts of the Rule in the Cache.
	// The Rule must have been executed already, both inputs and outputs must exist on disk.
	StoreArtifacts(*Rule) error
	// Fetch the output Artifacts of the Rule from the Cache and place them into the build directory.
	// Return an error if an Artifact failed to produce.
	ProcoduceArtifacts(*Rule) error
}

type PlanCache interface {
	// Return the plan.
	Plan() []string
	// Reset the plan.
	ClearPlan()
	// Schedule a Rule in the plan.
	ScheduleRule(*Rule)
}

type PrimordialCache interface {
	// Check if this (primordial) Rule is in the last run cache.
	// Return `true` if the Rule was run last run and its inputs didn't change since then. Return `false` otherwise.
	RuleInLastRun(*Rule) bool
	// Queue a new record of the Rule to to the last run cache.
	RecordRule(*Rule)
	// Refreshes the last run list of the cache. Old Rules are removed and new ones are added.
	// Return an error if it failed to atomically update the Cache.
	FlushRecords() error
}
