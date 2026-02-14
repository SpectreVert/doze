package doze

// Used to pass user configuration from the CLI to the Cache implementation.
type CacheOptions struct {
	Path string
}

// A Cache should implement all three nested interfaces.
//
// ArtifactCache is used to store and retrive Artifacts that have been produced in the current Doze run or a previous one.
//
// PrimordialCache is used to determine if a Rule needs to be rebuilt, that is if the content of its inputs has changed.
// It also records all the Rules of the Graph that was last traversed, regardless of them being executed this run.
//
// ArtifactCache and PrimordialCache must be persistent between invokations of Doze. Up to the implementer to decide how to store the data.
// Currently, there is no interface for cleaning up / rentention policies.
//
// PlanCache stores in RAM the `plan`, that is the list of Rules which have been resolved for execution.
// The plan is built by `graph.Resolve` and passed to `graph.Execute`, afterwhich it is cleared.
type Cache interface {
	ArtifactCache
	PrimordialCache
	PlanCache
}

// Manipulates and inspects output Artifacts.
type ArtifactCache interface {
	// Return `true` if *all* of the output Artifacts of the Rule are stored in the Cache, otherwise return `false`.
	HasArtifacts(*Rule) bool
	// Store the output Artifacts of the Rule in the Cache.
	// The Rule must have been executed already, both inputs and outputs must exist on disk.
	StoreArtifacts(*Rule) error
	// Fetch the output Artifacts of the Rule from the Cache and place them into the build directory.
	// Return an error if an Artifact failed to produce.
	ProduceArtifacts(*Rule) error
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

type PlanCache interface {
	// Return the plan.
	Plan() []string
	// Reset the plan.
	ClearPlan()
	// Schedule a Rule in the plan.
	ScheduleRule(*Rule)
}
