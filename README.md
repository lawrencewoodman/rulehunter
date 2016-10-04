Rulehuntersrv
=============
A server to find rules in data based on user specified goals

[![Build Status](https://travis-ci.org/vlifesystems/rulehuntersrv.svg?branch=master)](https://travis-ci.org/vlifesystems/rulehuntersrv)
[![Build status](https://ci.appveyor.com/api/projects/status/8tds5r4dk6163es0?svg=true)](https://ci.appveyor.com/project/LawrenceWoodman/rulehuntersrv)
[![Coverage Status](https://coveralls.io/repos/vlifesystems/rulehuntersrv/badge.svg?branch=master)](https://coveralls.io/r/vlifesystems/rulehuntersrv?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/vlifesystems/rulehuntersrv)](https://goreportcard.com/report/github.com/vlifesystems/rulehuntersrv)

Installation
------------
rulehuntersrv is compiled and installed from the root of the repository with:
```Shell
go install
```
There are several files needed to render the html properly these are located in the `support/` directory.

### Twitter Bootstrap

Copy directories from `support/bootstrap` to the `wwwDir` directory specified in `config.yaml`.

### jQuery

Copy directory `support/jquery/js` to the `wwwDir` directory specified in `config.yaml`.

### Html5 Shiv

Copy directory `support/html5shiv/js` to the `wwwDir` directory specified in `config.yaml`.

### Rulehuntersrv

Copy directories from `support/rulehuntersrv` to the `wwwDir` directory specified in `config.yaml`.

Usage
-----
rulehuntersrv is run using the `rulehuntersrv` executable created by `go install`.  You can use this command in a number of ways:

To processes the experiments in the `experimentsDir` directory specified in `config.yaml` located in the current directory:
```Shell
rulehuntersrv
```

To run `rulehuntersrv` as a server continually checking and processing experiments
```Shell
rulehuntersrv -serve
```

To install `rulehuntersrv` as a service (which then needs starting separately) with `config.yaml` located in `/usr/local/rulehuntersrv`:
```Shell
rulehuntersrv -install -configdir=/usr/local/rulehuntersrv
```

Configuration
-------------
rulehuntersrv is configured using a `config.yaml` file as follows:
```YAML
# The location of the experiment files
experimentsDir: "/usr/local/rulehuntersrv/experiments"
# The location of the html files produced
wwwDir: "/var/www/rulehuntersrv"
# The location of the build files created while running
buildDir: "/usr/local/rulehuntersrv/build"
# The source URL for the code to comply with the requirements of the AGPL
# (default: https://github.com/vlifesystems/rulehuntersrv)
sourceUrl: https://example.com/somecode/rulehuntersrv"
# The user to use when running as a service
user: rhuser
# The maximum number of rules in a report
# (default: 100)
maxNumReportRuels: 50
# The maximum number of processes used to process the experiments
# (default: -1, indicating the number of cpu's in the machine)
maxNumProcesses: 4
# The maximum number of records used from the data source.  This is
# useful when creating and testing experiment files
# (default: -1, indicating all the records)
maxNumRecords 150
```

Experiments
-----------
Each experiment is described by a `.yaml` or a `.json` file located in the `experimentsDir` of `config.yaml`.  This will look as follows:
```YAML
# The title of the report
title: "What would indicate good flow?"
# The tags associated with the report
tags:
  - test
  - "fred / ned"
# The type of the dataset (csv or sql)
dataset: "csv"
# Describe the CSV file
csv:
  # The name of the CSV file
  filename: "fixtures/flow.csv"
  # Whether the CSV file contains a header line
  hasHeader: true
  # What separator the CSV file is using
  separator:  ","
# The names of the fields to be used by rulehuntersrv
fieldNames:
  - group
  - district
  - height
  - flow
# Exclude the following fields from rule generation
excludeFieldNames:
  - flow
# Describe aggregators
aggregators:
    # The name of the aggregator
  - name: "goodFlowMCC"
    # The function to use (calc, count, mcc, mean, precision, recall, sum)
    function: "mcc"
    # The argument to pass to the aggregator function
    arg: "flow > 60"
# What goals to try to achieve
goals:
  - "goodFlowMCC > 0"
# The order to sort the rules in
sortOrder:
  - aggregatorName: "goodFlowMCC"
    direction: "descending"
  - aggregatorName: "numMatches"
    direction: "descending"
# When to run the experiment (default: !hasRun)
when: "!hasRunToday || sinceLastRunHours > 2"
```

### sql
To connect to a database using SQL you can replace the `csv` field above with something like this:
```YAML
sql:
  # The name of the driver to use (mssql, mysql, sqlite3)
  driverName: "mssql"
  # The details of the data source
  dataSourceName: "Server=127.0.0.1;Port=1433;Database=master;UID=sa,PWD=letmein"
  # An SQL query to run on the data source to create the dataset
  query: "select * from flow"
```

For more information about dataSourceName see the following for each driver:

* `mssql` - MS SQL Server - [README](https://github.com/denisenkom/go-mssqldb/blob/master/README.md).
* `mysql` - MySQL - [README](https://github.com/go-sql-driver/mysql#dsn-data-source-name).
* `sqlite3` - sqlite3 - This just uses the filename of the database.

<em>For security reasons any user specified for an SQL source should only have read access to the tables/database as the queries can't be checked for safety.</em>

### aggregators
The aggregators are used to collect data on the records that match against a rule.  There are the following functions:

* `calc` when supplied with an expression will calculate the result of that expression using as variables any aggregatorNames used in the expression.
* `count` will count the number of records that match a rule and the supplied expression.
* `mcc` calculates the [Matthews correlation coefficient](https://en.wikipedia.org/wiki/Matthews_correlation_coefficient) of a rule to match against the expression passed.  A coefficient of +1 represents a perfect prediction, 0 no better than random prediction and −1 indicates total disagreement between prediction and observation.
* `mean` calculates the mean value for an expression calculated on records that match a rule.
* `precision` calculates the [precision](https://en.wikipedia.org/wiki/Precision_and_recall) of a rule to match against the expression passed.
* `recall` calculates the [recall](https://en.wikipedia.org/wiki/Precision_and_recall) of a rule to match against the expression passed.
* `sum` calculates the sum for an expression calculated on records that match a rule.

### sortOrder
The rules in the report are sorted in the order of the entries for the `sortOrder` field.  The aggregators that can be used are the names specified in the `aggregators` field as well as the following built-in aggregators.

* `numMatches` is the number of records that a rule matches against.
* `percentMatches` is the percent of records in the dataset that a rule matches against.
* `goalsScore` reflects how well a rule passes the goals.  The higher the number the better the match.

### when
The `when` field determines when the experiment is to be run and how often.  It is an expression that supports the following variables:

* `hasRun` - whether the experiment has been ever been run
* `hasRunToday` - whether the experiment has been run today
* `hasRunThisWeek` - whether the experiment has been run this week
* `hasRunThisMonth` - whether the experiment has been run this month
* `hasRunThisYear` - whether the experiment has been run this year
* `sinceLastRunMinutes` - the number of minutes since the experiment was last run
* `sinceLastRunHours` - the number of hours since the experiment was last run
* `isWeekday` - whether today is a weekday

### expressions
Any expressions used in the experiment file conform to fairly standard Go expressions.

Testing
-------
To make testing simpler under Linux, where root is needed, you can use the following (replace _systemd_ with _upstart_ if using the latter init system):
```
sudo ./linux-test-su.sh $GOPATH `which go` systemd
```

To test the output of the server you can create a simple static webserver using something like the following from the `wwwDir` directory specified in the `config.yaml`:
```Shell
ruby -run -ehttpd . -p8000
```

If you don't like ruby there is this [list of one-liner static webservers](https://gist.github.com/willurd/5720255).

Requirements
------------
These have all be vendored into the `vendor/` directory using the
[govendor](https://github.com/kardianos/govendor) tool.

* [dexpr](https://github.com/lawrencewoodman/dexpr) package
* [dlit](https://github.com/lawrencewoodman/dlit) package
* [rulehunter](https://github.com/vlifesystems/rulehunter) package
* [osext](https://github.com/kardianos/osext) package
* [service](https://github.com/kardianos/service) package
* [go-mssqldb](https://github.com/denisenkom/go-mssqldb) package
* [go-sqlite3](https://github.com/mattn/go-sqlite3) package
* [mysql](https://github.com/go-sql-driver/mysql) package
* [yaml.v2](https://gopkg.in/yaml.v2) package


To format the html pages properly [Twitter Boostrap](http://getbootstrap.com) is used and hence has been included in the `support/` directory.

Contributing
------------
If you want to improve this program make a pull request to the [repo](https://github.com/vlifesystems/rulehuntersrv) on github.  Please put any pull requests in a separate branch to ease integration and add a test to prove that it works.  If you find a bug, please report it at the project's [issues tracker](https://github.com/vlifesystems/rulehuntersrv/issues) also on github.

Licence
-------
Rulehuntersrv - A server to find rules in data based on user specified goals

Copyright (C) 2016 [vLife Systems Ltd](http://vlifesystems.com)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program; see the file COPYING.  If not, see
[http://www.gnu.org/licenses/](http://www.gnu.org/licenses/).

Additional Licences
-------------------

### Twitter Boostrap

[Twitter Boostrap](http://getbootstrap.com) has been included in the repository under `support/bootstrap`.

Copyright 2011-2015 Twitter, Inc.

Licensed under the MIT license.  For details see: [http://getbootstrap.com](http://getbootstrap.com).

### jQuery

[jQuery](https://jquery.org) has been included in the repository under `support/jquery`.

Copyright jQuery Foundation

Licensed under the MIT license.  For details see [https://jquery.org/license/](https://jquery.org/license/).

### Html5 Shiv

[Html5 Shiv](https://github.com/aFarkas/html5shiv) has been included in part in the repository under `support/html5shiv`

Copyright (c) 2014 Alexander Farkas (aFarkas).

This is dual licensed under the MIT and GPL version 2 licence.  For the sake of Rulehuntersrv we will take it to be the MIT license as this makes it easier to combine with the Affero GPL version 3 license that Rulehuntersrv is licenced under.  For details see the licence file in `support/html5shiv`.

### Respond

[Respond](https://github.com/scottjehl/Respond) has been included in part in the repository under `support/respond`

Copyright (c) 2013 Scott Jehl

Licensed under the MIT license.  For details see the license file in `support/respond`.

### Vendored Go Packages

This repository includes other packages in the `vendor/` directory.  Please see those packages for the licences that cover those works.
