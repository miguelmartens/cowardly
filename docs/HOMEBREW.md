# Homebrew tap — version ldflag

The binary shows its version with `cowardly version` or `cowardly -v`. To avoid showing `dev`, the **Homebrew formula must inject the version** at build time.

In the tap repo [miguelmartens/homebrew-cowardly](https://github.com/miguelmartens/homebrew-cowardly), ensure the formula passes the version in ldflags:

```ruby
def install
  ldflags = "-s -w -X main.Version=v#{version}"
  system "go", "build", *std_go_args(ldflags: ldflags), "./cmd/cowardly"
end
```

So the built binary reports e.g. `Cowardly version: v0.2.5` instead of `Cowardly version: dev`. Homebrew’s `version` is taken from the tag in the tarball URL (e.g. `v0.2.5` → `version` is `0.2.5`), hence `v#{version}`.
