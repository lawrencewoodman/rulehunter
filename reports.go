/*
 * Copyright (C) 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
 */

package main

import (
	"errors"
	"fmt"
	"github.com/lawrencewoodman/dexpr_go"
	"github.com/lawrencewoodman/dlit_go"
	"github.com/lawrencewoodman/rulehunter"
	"html/template"
	"os"
	"path/filepath"
	"sort"
)

func writeReport(
	assessment *rulehunter.Assessment,
	experiment *rulehunter.Experiment,
	experimentFilename string,
	reportsDir string,
) error {
	assessment.Sort(experiment.SortOrder)
	assessment.Refine(1)
	const tpl = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>{{.Title}}</title>
	</head>

	<style>
		table {
			border-collapse: collapse;
		}

		table th {
			text-align: left;
			padding-left: 1em;
			padding-right: 2em;
			border-collapse: collapse;
			border-right: 1px solid black;
			border-bottom: 1px solid black;
		}

		table th.last-column {
			border-right: 0;
		}
		table td {
			border-collapse: collapse;
			border-right: 1px solid black;
			padding-left: 1em;
			padding-right: 1em;
		  padding-top: 0.1em;
		  padding-bottom: 0.1em;
		}
		table td.last-column {
			border-right: 0;
		}
		table tr.title td {
			border-bottom: 1px solid black;
		}
		div {
			margin-bottom: 2em;
		}
		div.aggregators {
			float: left;
		  clear: left;
			margin-right: 3em;
		}
		div.goals {
			float: left;
		  clear: right;
		}

		div.config table {
			margin-bottom: 2em;
		}
		div.config table th {
			height: 1em;
		}
	</style>

	<body>
		<h1>{{.Title}}</h1>

		<div class="config">
			<h2>Config</h2>
			<table>
				<tr class="title">
					<th> </th>
					<th class="last-column"> </th>
				</tr>
				<tr>
					<td>Number of records processed</td>
					<td class="last-column">{{.NumRecords}}</td>
				</tr>
				<tr>
					<td>Experiment file</td>
					<td class="last-column">{{.ExperimentFilename}}</td>
				</tr>
			</table>

			<table>
				<tr class="title">
					<th>Sort Order</th><th class="last-column">Direction</th>
				</tr>
				{{range .SortOrder}}
					<tr>
						<td>{{ .Field }}</td><td class="last-column">{{ .Direction }}</td>
					</tr>
				{{end}}
			</table>
		</div>

		<h2>Results</h2>
		{{range .Assessments}}
			<h3 style="clear: both;">{{ .Rule }}</h3>


			<div class="aggregators">
				<table>
					<tr class="title">
						<th>Aggregator</th>
            <th>Value</th>
						<th class="last-column">Improvement</th>
					</tr>
					{{ range .Aggregators }}
					<tr>
						<td>{{ .Name }}</td>
						<td>{{ .Value }}</td>
						<td class="last-column">{{ .Difference }}</td>
					</tr>
					{{ end }}
				</table>
			</div>

			<div class="goals">
				<table>
					<tr class="title"><th>Goal</th><th class="last-column">Value</th></tr>
					{{ range $key, $value := .Goals }}
					<tr>
						<td>{{ $key }}</td><td class="last-column">{{ $value }}</td>
					</tr>
					{{ end }}
				</table>
			</div>
		{{ end }}
	</body>
</html>`

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		return err
	}

	type TplAggregator struct {
		Name       string
		Value      string
		Difference string
	}

	type TplAssessment struct {
		Rule        string
		Aggregators []*TplAggregator
		Goals       map[string]bool
	}

	type TplData struct {
		Title              string
		ExperimentFilename string
		NumRecords         int64
		SortOrder          []rulehunter.SortField
		Assessments        []*TplAssessment
	}

	trueAggregators, err := getTrueAggregators(assessment)
	if err != nil {
		return err
	}

	assessments := make([]*TplAssessment, len(assessment.RuleAssessments))
	for i, ruleAssessment := range assessment.RuleAssessments {
		aggregatorNames := getSortedAggregatorNames(ruleAssessment.Aggregators)
		aggregators := make([]*TplAggregator, len(ruleAssessment.Aggregators))
		j := 0
		for _, aggregatorName := range aggregatorNames {
			aggregator := ruleAssessment.Aggregators[aggregatorName]
			difference :=
				calcTrueAggregatorDifference(trueAggregators, aggregator, aggregatorName)
			aggregators[j] = &TplAggregator{
				aggregatorName,
				aggregator.String(),
				difference,
			}
			j++
		}
		assessments[i] = &TplAssessment{
			ruleAssessment.Rule.String(),
			aggregators,
			ruleAssessment.Goals,
		}
	}
	tplData := TplData{
		experiment.Title,
		experimentFilename,
		assessment.NumRecords,
		experiment.SortOrder,
		assessments,
	}

	outputFilename :=
		filepath.Join(reportsDir, fmt.Sprintf("%s.html", experimentFilename))
	f, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, tplData)
}

func getSortedAggregatorNames(aggregators map[string]*dlit.Literal) []string {
	aggregatorNames := make([]string, len(aggregators))
	j := 0
	for aggregatorName, _ := range aggregators {
		aggregatorNames[j] = aggregatorName
		j++
	}
	sort.Strings(aggregatorNames)
	return aggregatorNames
}

func getTrueAggregators(
	assessment *rulehunter.Assessment,
) (map[string]*dlit.Literal, error) {
	trueRuleAssessment :=
		assessment.RuleAssessments[len(assessment.RuleAssessments)-1]
	if trueRuleAssessment.Rule.String() != "true()" {
		return map[string]*dlit.Literal{}, errors.New("Can't find true() rule")
	}
	trueAggregators := trueRuleAssessment.Aggregators
	return trueAggregators, nil
}

func calcTrueAggregatorDifference(
	trueAggregators map[string]*dlit.Literal,
	aggregatorValue *dlit.Literal,
	aggregatorName string,
) string {
	diffExpr, err := dexpr.New("r - t")
	if err != nil {
		panic(fmt.Sprintf("Couldn't create difference expression: %s", err))
	}
	funcs := map[string]dexpr.CallFun{}
	vars := map[string]*dlit.Literal{
		"r": aggregatorValue,
		"t": trueAggregators[aggregatorName],
	}
	difference := "N/A"
	differenceL := diffExpr.Eval(vars, funcs)
	if !differenceL.IsError() {
		difference = differenceL.String()
	}
	return difference
}
