### Master branch

  * Switch to MIT Licence
  * s/`DescribingError`/`DescribeError`/
  * Correct MergeError error message

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
