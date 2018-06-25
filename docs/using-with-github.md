## Using Gitkube with external git providers (like Github)

As it stands, a Gitkube remote is just for deployment and should not be considered as a storage medium like GitHub.
If however one would like for `git push` to both deploy to a kubernetes cluster as well as store code for the long haul in GitHub,
this can be achieved.

One simple way to achieve this is to begin with an origin remote pointing to Github and then add the Gitkube repository as a push
url for the origin remote. The following example presumes that you already have two remotes: one pointing to Github and the other
to Gitkube.

```
$ git remote -v
origin  git@github.com/<github-owner>:<github-repo>.git (fetch)
origin  git@github.com/<github-owner>:<github-repo>.git (push)
myremote	<gitkube-remote-url> (fetch)
myremote	<gitkube-remote-url> (push)
```

Now let's set both urls for the `origin` remote.

```
$ git remote set-url --add --push origin git@github.com/<github-owner>:<github-repo>.git
$ git remote set-url --add --push origin <gitkube-remote-url>
$ git remote -v
origin git@github.com/<github-owner>:<github-repo>.git (fetch)
origin git@github.com/<github-owner>:<github-repo>.git (push)
origin <gitkube-remote-url> (push)
myremote	<gitkube-remote-url> (fetch)
myremote	<gitkube-remote-url> (push)
```

Now `git push origin master` will push the code to both remotes.

Fore more info check out [this Stack Overflow thread.](https://stackoverflow.com/questions/14290113/git-pushing-code-to-two-remotes)
