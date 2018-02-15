Contributing to Rulehunter
==========================

We would love contributions to improve this project.  You can help by reporting bugs, improving the documentation, submitting feature requests, fixing bugs, etc.

Table of Contents
-----------------

  * [Code](#code)
    * [Reporting Issues](#reporting-issues)
    * [Code Contributions](#code-contributions)
    * [Licence](#licence)
    * [Additional Licences and attribution](#additional-licences-and-attribution)
  * [Documentation](#documentation)

Code
----
Rulehunter is written in [Go](https://golang.org/) (v1.8+) and the code can be fetched with:

    go get github.com/vlifesystems/rulehunter

Rulehunter uses [govendor](https://github.com/kardianos/govendor) to vendor dependencies and these are included in the repository's `vendor/` directory.

### Reporting Issues
If you find a an issue with the code, please report it using the project's [issue tracker](https://github.com/vlifesystems/rulehunter/issues).  When reporting the issue please provide the version of Rulehunter (`rulehunter version`) and operating system being used.

### Code Contributions

To submit code contributions we ask the following:

  * Confirm that your contributions can be included via a simple [Developer Certificate of Origin](#developer-certificate-of-origin) which we use instead of a full CLA. This can be signed via [cla-assistant](https://cla-assistant.io/vlifesystems/rulehunter).
  * Please put your changes in a separate branch to ease integration.
  * For new code please add tests to prove that it works.
  * Update [CHANGELOG.md](https://github.com/vlifesystems/rulehunter/blob/master/CHANGELOG.md) if appropriate.
  * Make a pull request to the [repo](https://github.com/vlifesystems/rulehunter) on github.

#### Developer Certificate of Origin
Rather than use a full-blown _Contributor License Agreement_, this project uses a simple _Developer Certificate of Origin_ to confirm that any contributions can be included.  The text of the certificate is listed below and the easiest way to confirm this is via [cla-assistant](https://cla-assistant.io/vlifesystems/rulehunter).

    Developer Certificate of Origin
    Version 1.1

    Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
    1 Letterman Drive
    Suite D4700
    San Francisco, CA, 94129

    Everyone is permitted to copy and distribute verbatim copies of this
    license document, but changing it is not allowed.


    Developer's Certificate of Origin 1.1

    By making a contribution to this project, I certify that:

    (a) The contribution was created in whole or in part by me and I
        have the right to submit it under the open source license
        indicated in the file; or

    (b) The contribution is based upon previous work that, to the best
        of my knowledge, is covered under an appropriate open source
        license and I have the right under that license to submit that
        work with modifications, whether created in whole or in part
        by me, under the same open source license (unless I am
        permitted to submit under a different license), as indicated
        in the file; or

    (c) The contribution was provided directly to me by some other
        person who certified (a), (b) or (c) and I have not modified
        it.

    (d) I understand and agree that this project and the contribution
        are public and that a record of the contribution (including all
        personal information I submit with it, including my sign-off) is
        maintained indefinitely and may be redistributed consistent with
        this project or the open source license(s) involved.


### Licence
Copyright (C) 2016-2018 [vLife Systems Ltd](http://vlifesystems.com)

Rulehunter is licensed under an MIT licence.  Please see [LICENSE.md](https://github.com/vlifesystems/rulehunter/blob/master/LICENSE.md) for details.

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

This is dual licensed under the MIT and GPL version 2 licence.  For the sake of Rulehunter we will take it to be the MIT license as this matches the type of licence used by Rulehunter.  For details see the licence file in `support/html5shiv`.

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
