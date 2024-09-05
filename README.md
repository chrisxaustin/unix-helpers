[![tests](https://github.com/chrisxaustin/unix-helpers/actions/workflows/build-main.yml/badge.svg)](https://github.com/ForesiteMSSP/chronicle_controllers_automation/actions/workflows/test.yml)

# unix-tools

A collection of unix command-line utilities.

# tf

Functions similarly to `tail -F`, but with an idle callback.

* allows tailing multiple files
* continues to follow files after they are rotated
* continues to follow files if they are deleted and recreated
* prints a line of dashes after 5s of no activity, which helps when watching for new content
