# snap plugin collector docker

1. [Contributing Code](#contributing-code)
2. [Contributing Examples](#contributing-examples)
3. [Contribute Elsewhere](#contribute-elsewhere)
4. [Thank You](#thank-you)

This repository has dedicated developers from Intel working on updates. The most helpful way to contribute is by reporting your experience through issues. Issues may not be updated while we review internally, but they're still incredibly appreciated.

## Contributing Code
**_IMPORTANT_**: We encourage contributions to the project from the community. We ask that you keep the following guidelines in mind when planning your contribution.

* Whether your contribution is for a bug fix or a feature request, **create an [Issue](https://github.com/intelsdi-x/snap-plugin-collector-docker/issues)** and let us know what you are thinking.
* **For bugs**, if you have already found a fix, feel free to submit a Pull Request referencing the Issue you created. Include the `Fixes #` syntax to link it to the issue you're addressing.
* **For feature requests**, we want to improve upon the library incrementally which means small changes at a time. In order to ensure your PR can be reviewed in a timely manner, please keep PRs small, e.g. <10 files and <500 lines changed. If you think this is unrealistic, then mention that within the issue and we can discuss it.

Once you're ready to contribute code back to this repo, start with these steps:

* Fork the appropriate sub-projects that are affected by your change.
* Clone the fork to `$GOPATH/src/github.com/intelsdi-x/`:

    ```
$ git clone https://github.com/<yourGithubID>/<project>.git
    ```
* Create a topic branch for your change and checkout that branch:

    ```
    $ git checkout -b some-topic-branch
    ```
* Make your changes and run the test suite if one is provided.
* Commit your changes and push them to your fork.
* Open a pull request for the appropriate project.
* Contributors will review your pull request, suggest changes, and merge it when it’s ready and/or offer feedback.

If you have questions feel free to contact the [maintainers](https://github.com/intelsdi-x/snap/blob/master/README.md#maintainers) by tagging them: @intelsdi-x/plugin-maintainers.

## Contributing Examples
The most immediately helpful way you can benefit this project is by cloning the repository, adding some further examples and submitting a pull request.

Have you written a blog post about how you use [Snap](http://github.com/intelsdi-x/snap) and/or this plugin? Send it to us [on Slack](http://slack.snap-telemetry.io)!

## Contribute Elsewhere
This repository is one of **many** plugins in **Snap**, a powerful telemetry framework. See the full project at http://snap-telemetry.io

## Thank You
And **thank you!** Your contribution, through code and participation, is incredibly important to us.
