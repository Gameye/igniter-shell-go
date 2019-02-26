# Gameye Shell

## Automated Tests
NEVER commit something that breaks the build! Please link the `test.sh` script
as a git `pre-commit` hook so you are shure that tests are run before every
commit.

like this:
```bash
ln test.sh .git/hooks/pre-commit
```
