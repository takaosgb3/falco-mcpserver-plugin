# Release Checklist

This checklist ensures all necessary steps are completed before releasing a new version.

## Pre-Release Checks

### 1. Code Quality
- [ ] All tests pass locally
- [ ] Linting passes (`make lint`)
- [ ] Build succeeds (`make build`)

### 2. Rule Validation
- [ ] Validate Falco rules locally or on EC2:
  ```bash
  sudo falco --validate falco-plugin-nginx-public/rules/nginx_rules.yaml
  ```
- [ ] Test rules with actual nginx logs
- [ ] Verify no parentheses with commas in output fields
- [ ] Ensure all rules have `source: nginx`

### 3. Version Update
- [ ] Run version update script:
  ```bash
  ./scripts/update-version.sh v0.4.X
  ```
- [ ] Verify all documentation updated:
  - [ ] `falco-plugin-nginx-public/docs/QUICK_START_BINARY_INSTALLATION.md`
  - [ ] `falco-plugin-nginx-public/README.md`
  - [ ] `falco-plugin-nginx-public/CHANGELOG.md`
  - [ ] Rule file version comments

### 4. Testing
- [ ] Test installation script on fresh EC2 instance:
  ```bash
  curl -sSL https://raw.githubusercontent.com/takaosgb3/falco-plugin-nginx/main/install.sh | sudo bash
  ```
- [ ] Verify plugin loads correctly
- [ ] Test attack detection works

## Release Process

### 1. Commit Changes
```bash
git add -A
git commit -m "chore: prepare release vX.X.X"
git push origin main
```

### 2. Create Tag
```bash
git tag -a vX.X.X -m "Release vX.X.X - <brief description>"
git push origin vX.X.X
```

### 3. Create GitHub Release
```bash
gh release create vX.X.X \
  --title "Release vX.X.X - <title>" \
  --notes "<release notes>" \
  /path/to/libfalco-nginx-plugin-linux-amd64.so \
  rules/nginx_rules.yaml
```

### 4. Generate and Upload Checksums
```bash
sha256sum libfalco-nginx-plugin-linux-amd64.so rules/nginx_rules.yaml > checksums.txt
gh release upload vX.X.X checksums.txt
```

## Post-Release

### 1. Verification
- [ ] Test installation from release:
  ```bash
  PLUGIN_VERSION=vX.X.X curl -sSL https://raw.githubusercontent.com/takaosgb3/falco-plugin-nginx/main/install.sh | sudo bash
  ```
- [ ] Verify checksums match
- [ ] Test on multiple platforms if possible

### 2. Communication
- [ ] Update any external documentation
- [ ] Notify users if breaking changes

## Common Issues to Avoid

1. **Rule Format Errors**
   - No parentheses with commas in output
   - Use space-separated format: `field=%value% field2=%value2%`
   - All rules must have `source: nginx`

2. **Version Consistency**
   - Always use `update-version.sh` script
   - Check all documentation files
   - Update CHANGELOG.md

3. **Binary Compatibility**
   - Build on Linux (not macOS)
   - Use proper build flags: `-buildmode=c-shared`
   - Verify with `file` command (should be "ELF 64-bit LSB shared object")

## Automation Opportunities

Consider implementing:
1. GitHub Actions workflow for automated releases
2. Pre-commit hooks for version consistency
3. Automated testing on multiple platforms
4. Release notes generation from commits