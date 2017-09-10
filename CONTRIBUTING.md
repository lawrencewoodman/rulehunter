Contributing to Rulehunter
==========================

We would love contributions to improve this project.  You can help by reporting bugs, improving the documentation, submitting feature requests, fixing bugs, etc.

Table of Contents
-----------------

  * [Code](#code)
    * [Reporting Issues](#reporting-issues)
    * [Licence](#licence)
    * [Additional Licences and attribution](#additional-licences-and-attribution)
  * [Documentation](#documentation)

Code
----
Rulehunter is written in [Go](https://golang.org/) (v1.8+) and the code can be fetched with:

    go get github.com/vlifesystems/rulehunter

To submit code contributions we ask the following:

  * Please put your changes in a separate branch to ease integration.
  * For new code please add tests to prove that it works.
  * Update [CHANGELOG.md](https://github.com/vlifesystems/rulehunter/blob/master/CHANGELOG.md) if appropriate.
  * Make a pull request to the [repo](https://github.com/vlifesystems/rulehunter) on github.

### Vendored Dependencies

Rulehunter uses [govendor](https://github.com/kardianos/govendor) to vendor dependencies and these are included in the repository's `vendor/` directory.

### Reporting Issues
If you find a an issue with the code, please report it using the project's [issue tracker](https://github.com/vlifesystems/rulehunter/issues).  When reporting the issue please provide the version of Rulehunter (`rulehunter version`) and operating system being used.

### Licence
Rulehunter - Find rules in data based on user specified goals

Copyright (C) 2016-2017 [vLife Systems Ltd](http://vlifesystems.com)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program; see the file [COPYING](https://github.com/vlifesystems/rulehunter/blob/master/COPYING).  If not, see
[http://www.gnu.org/licenses/](http://www.gnu.org/licenses/).

### Additional Licences and Attribution

#### Twitter Boostrap

[Twitter Boostrap](http://getbootstrap.com) has been included in the repository under `support/bootstrap`.

Copyright 2011-2015 Twitter, Inc.

Licensed under the MIT license.  For details see: [http://getbootstrap.com](http://getbootstrap.com).

#### jQuery

[jQuery](https://jquery.org) has been included in the repository under `support/jquery`.

Copyright jQuery Foundation

Licensed under the MIT license.  For details see [https://jquery.org/license/](https://jquery.org/license/).

#### Html5 Shiv

[Html5 Shiv](https://github.com/aFarkas/html5shiv) has been included in part in the repository under `support/html5shiv`

Copyright (c) 2014 Alexander Farkas (aFarkas).

This is dual licensed under the MIT and GPL version 2 licence.  For the sake of Rulehunter we will take it to be the MIT license as this makes it easier to combine with the Affero GPL version 3 license that Rulehunter is licenced under.  For details see the licence file in `support/html5shiv`.

#### Respond

[Respond](https://github.com/scottjehl/Respond) has been included in part in the repository under `support/respond`

Copyright (c) 2013 Scott Jehl

Licensed under the MIT license.  For details see the license file in `support/respond`.

#### loading.io

A loading icon called `ring.gif` has been included under `support/rulehunter/img`.  This came from the [loading.io](http://loading.io) website and their terms of uses state:

    All materials used in generating animated icons in this website, except the g0v icon, are created by loading.io. You can use them freely for any purpose.


#### Vendored Go Packages

This repository includes other packages in the `vendor/` directory.  Please see those packages for the licences that cover those works.


Documentation
-------------
Documentation is really important to us and is kept centrally at [rulehunter.com](http://rulehunter.com). The content is licensed under a [Creative Commons Attribution 4.0 International License](http://creativecommons.org/licenses/by/4.0/).

If you want to help with this make a pull request to the [website repo](https://github.com/vlifesystems/rulehunter.com) on github. Please put any pull requests in a separate branch to ease integration.

### Reporting Issues
If you find an issue with the documentation, please report it using the website's [issue tracker](https://github.com/vlifesystems/rulehunter.com/issues).
