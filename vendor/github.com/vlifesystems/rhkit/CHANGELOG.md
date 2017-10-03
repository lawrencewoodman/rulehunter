## Master

  * Switch to MIT Licence
  * s/`DescribingError`/`DescribeError`/
  * s/`aggregators`/`aggregator`/ for package name
  * Create `aggregator.MakeSpecs`
  * s/`aggregator.AggregatorSpec`/`aggregator.Spec`/
  * s/`aggregator.AggregatorInstance`/`aggregator.Instance`/
  * Create `rule.MakeDynamicRules` and allow these to be passed to `Process`
  * Create `goal.MakeGoals`
  * Move sort descriptions to `assessment`
  * Add `FieldNames` method to `Description`
  * Move `experiment.AggregatorDesc` to `aggregator.Desc` and rename
    its `Function` variable to `Kind`
  * Move aggregator specific errors from `experiment` to a consolidated
    `DescError` type in `aggregator`
  * Remove `experiment` package
  * Create `GenerateRulesError`
  * Create `Options` struct to pass some parameters to `Process`
  * Unexport `description.New` and `Description.NextRecord` method
  * Change error for `Assessment.Merge`
  * Create `Assessment.AssessRules` method to replace `assessment.AssessRules`
    function
  * Remove `experiment.Experiment`
  * Change `Assessment.AssessRules` and `Process` so that an `Experiment` is
    no longer passed in, instead the needed fields from the `Experiment` are
    passed
  * Consolidate `assessment.RuleAssessment` and `assessment.RuleAssessor` into
    `assessment.RuleAssessment`
  * Change fields in `Options` struct and therefore `Process` function
  * Have `rule.Generate` return an error if rule fields not valid
  * Change `rule.Generate` to use `GenerationDescriber` interface

## 0.2 (15th July 2017)

  * Rename rulehunter to rhkit
  * Unexport `Flags` in `Assessment` struct
  * s/`LimitRuleAssessments`/`TruncateRuleAssessments`/
  * s/`numGoalsPassed`/`goalsScore`/
  * Use external package ddataset
  * Add `mcc`, `mean`, `precision` and `recall` aggregators
  * Restructure aggregators to use a registration system
  * Switch from dynamic rules to static
  * Remove `AssessRulesMP`
  * Remove `accuracy` and `percent` aggregators
  * Remove 100* from `percentMatches` aggregator result
  * Remove `NI` rules
  * Use `ruleFields` rather than `excludeFields`
  * Add `pow`, `min`, `max` to `dexprfuncs`
  * Record number of each value in dataset description
  * Create `rules.Combine` to combine rules
  * Create `rule.Tweaker`/`Overlapper`/`DPReducer`/`Valuer` interfaces
  * Add  `WriteJSON`/`LoadJSON` for `Description`
  * Add arithmetic rules such as `a+b>=c`
  * Add `rule.ReduceDP`
  * Move rule generation to `rule.Generate`
  * Move `DescribeDataset` into `Description`
  * Create `assessment` sub-package
  * Create Process function
  * Up Go requirement to v1.7


## 0.1 (21st May 2016)

 * Initial release of rulehunter
