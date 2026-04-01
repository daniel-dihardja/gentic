package reflect

// ReflectUILabelTotal returns the denominator for "Refining (i/total)" in the activity feed.
// It matches apps/gentic-agents config MaxReflectionIterations / MAX_REFLECTION_ITERATIONS:
// 0 or 1 means a single user-facing reflection step (shows 1/1, not 1/2 when max is 1).
// For 2 or more, total is maxReflection+1 (one initial draft plus up to maxReflection revisions).
func ReflectUILabelTotal(maxReflection int) int {
	if maxReflection <= 1 {
		return 1
	}
	return maxReflection + 1
}

func reflectUILabelPair(iteration, maxReflection int) (current, total int) {
	total = ReflectUILabelTotal(maxReflection)
	if maxReflection <= 1 {
		return 1, total
	}
	return iteration + 1, total
}
