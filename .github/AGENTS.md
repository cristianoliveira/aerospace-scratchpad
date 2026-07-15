# .github

GitHub Actions and issue templates.

- Reuse Make targets so local and CI verification stay aligned.
- Keep release, version bump, nightly reset, and push validation workflows separate.
- Pin action versions deliberately and minimize workflow permissions.
- Preserve required secrets and release inputs; never embed credentials.
- Validate YAML and inspect trigger/path changes carefully because they alter when automation runs.
