## Master

### Config

 * Remove sourceURL

### Experiment Files

 * s/`function`/`kind`/ for `aggregators` in experiment files
 * s/`aggregatorName`/`aggregator`/ for `sortOrder` in experiment files
 * Add `rules` to experiment files to allow user defined rules
 * Add `category` to experiment files
 * Replace `complexity` and `ruleFields` in experiment files
   with `ruleGeneration`

### HTML

 * Fix support for HTML5 in IE8
 * Change directory structure of generated html
 * Create dashboard style front page for generated html
 * s/failure/error/ in activity
 * Remove 'licence' page in generated html

### CLI

 * Specify config filename rather than config directory through CLI
 * Change CLI to use a command structure. E.G. `rulehunter serve --config=.`
 * Move `user` from config file to CLI flag for `service`
 * Add `version` command to CLI

### Build

 * Restructure `progress.json` file
 * Put dataset 'Description' into 'Report'
 * Use sha-512 hash to create 'reports' filenames

### Development

 * Switch to MIT licence
 * Up Go requirement to v1.8+

## 0.1 (7th July 2017)

 * Initial release of Rulehunter
