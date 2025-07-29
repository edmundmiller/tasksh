# Installing

## Using Homebrew (macOS/Linux)
```bash
brew install edmundmiller/taskagent/taskagent
```

Or tap the repository first:
```bash
brew tap edmundmiller/taskagent
brew install taskagent
```

## Building from source
```bash
go install github.com/emiller/tasksh/cmd/tasksh@latest
```

## Package managers
* Debian/Ubuntu:
```
$ sudo apt-get install tasksh
```

# Disclaimer during ongoing development

The development branch is a work in progress and may not pass all quality tests,
therefore it may harm your data. While we welcome bug reports from the
development branch, we do not guarantee proper or timely fixes.

- Make proper backups.
- Broken functionality may arise from ongoing development work.
- Be aware that using the development branch involves risks.

---

Thank you for taking a look at tasksh!!

---

Tasksh is released under the MIT license. For details check the LICENSE file.

# Important note
When cloning this from the repo to build from source make sure you `git clone --recursive` to get all required submodules.
