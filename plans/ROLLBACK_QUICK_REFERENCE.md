# Rollback Quick Reference

**For detailed instructions, see: [ROLLBACK_PLAN.md](ROLLBACK_PLAN.md)**

---

## Quick Decision Guide

### No Implementation Yet → Type A (1 minute)
```bash
rm -rf plans/  # Just delete the plans directory
```
**Git Required:** No

---

### Uncommitted Changes → Type B (5-10 minutes)
```bash
git stash push -m "Backup multi-profile work"  # Backup
git restore .                                   # Restore files
git clean -fd                                   # Remove new files
rm -f windows/profileManager.go windows/profileManagerDialog.go
```
**Git Required:** Yes

---

### Local Commits Only → Type C (15-30 minutes)
```bash
git branch backup/multi_profile_impl           # Backup
git reset --hard <commit-before-implementation> # Reset
rm -f windows/profileManager.go windows/profileManagerDialog.go
```
**Git Required:** Yes

---

### Pushed to Remote → Type D (30-60 minutes)
```bash
git checkout -b rollback/multi_profile
git revert <commit1> <commit2> ...             # Revert in reverse order
rm -f windows/profileManager.go windows/profileManagerDialog.go
git push origin rollback/multi_profile
```
**Git Required:** Yes

---

## Files to Delete

### New Files Created:
```
windows/profileManager.go
windows/profileManagerDialog.go
plans/*.md (all plan files)
```

### Modified Files to Restore:
```
windows/mainWindow.go    - Major changes
windows/fileDialog.go    - Minor changes
```

---

## Verification Commands

```bash
# Check status
git status
git diff

# Test build
go build
./dsb

# Verify git history
git log --oneline -10
```

---

## Recovery If Needed Later

```bash
# If you created a backup:
git merge backup/multi_profile_impl

# Or cherry-pick from history:
git cherry-pick <commit-hash>

# Or re-implement using the plans
```

---

## Current Status (2025-10-17)

- **Plans Directory:** Untracked (not in git)
- **Implementation Status:** Not started
- **Recommended Rollback:** Type A (simple delete)
- **Git Required:** No

---

## Emergency Quick Rollback

```bash
# If application is broken in production:
git checkout <last-known-good-commit>
go build
# Deploy immediately

# Then plan proper rollback during maintenance
```

---

**For full details, decision trees, and step-by-step instructions:**
**See: [ROLLBACK_PLAN.md](ROLLBACK_PLAN.md)**
