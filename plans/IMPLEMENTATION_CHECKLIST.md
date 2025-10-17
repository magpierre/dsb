# Implementation Checklist - Multi-Profile Support

**Purpose:** Track progress and ensure safe implementation with rollback points

---

## Pre-Implementation Setup

- [ ] Read all plan documents (README.md, STEP1-7)
- [ ] Review ROLLBACK_PLAN.md
- [ ] Understand current git status
- [ ] **Create feature branch:**
  ```bash
  git checkout -b feature/multi_profile
  ```
- [ ] **Create initial backup:**
  ```bash
  git branch backup/before_multi_profile_$(date +%Y%m%d)
  ```

---

## Step 1: Foundation (2-3 hours)

### Pre-Step Checkpoint:
- [ ] Read STEP1_FOUNDATION.md completely
- [ ] Current working directory clean (`git status`)
- [ ] Application builds and runs

### Implementation Tasks:
- [ ] Create `windows/profileManager.go`
- [ ] Add ProfileInfo struct
- [ ] Add ProfileManager struct with methods:
  - [ ] NewProfileManager()
  - [ ] AddProfile()
  - [ ] GetProfile()
  - [ ] ListProfiles()
  - [ ] RemoveProfile()
  - [ ] SetActiveProfile()
  - [ ] GetActiveProfile()
- [ ] Update MainWindow struct in `windows/mainWindow.go`
- [ ] Update Selected struct in `windows/mainWindow.go`

### Testing:
- [ ] Code compiles (`go build`)
- [ ] No new runtime errors
- [ ] ProfileManager can be instantiated

### Post-Step Actions:
- [ ] **Commit changes:**
  ```bash
  git add windows/profileManager.go windows/mainWindow.go
  git commit -m "Step 1: Add ProfileManager foundation"
  ```
- [ ] **Tag checkpoint:**
  ```bash
  git tag checkpoint/step1
  ```

### Rollback Point:
```bash
# If Step 1 fails:
git reset --hard HEAD~1
rm -f windows/profileManager.go
```

---

## Step 2: Profile Loading (2-3 hours)

### Pre-Step Checkpoint:
- [ ] Step 1 completed successfully
- [ ] Read STEP2_PROFILE_LOADING.md completely
- [ ] Application still runs normally

### Implementation Tasks:
- [ ] Update `OpenProfile()` in mainWindow.go
- [ ] Add `showProfileLoadModeDialog()`
- [ ] Add `loadProfileAndAdd()`
- [ ] Add `replaceAllProfiles()`
- [ ] Add `loadSharesForProfile()`
- [ ] Add `switchToProfile()`
- [ ] Add `initializeClient()` to ProfileManager
- [ ] Update fileDialog.go if needed

### Testing:
- [ ] Can load first profile
- [ ] Dialog appears when loading second profile
- [ ] "Add to Existing" works
- [ ] "Replace All" works
- [ ] Shares load correctly for each profile

### Post-Step Actions:
- [ ] **Commit changes:**
  ```bash
  git add windows/mainWindow.go windows/fileDialog.go
  git commit -m "Step 2: Add profile loading with user choice dialog"
  git tag checkpoint/step2
  ```

### Rollback Point:
```bash
# If Step 2 fails:
git reset --hard checkpoint/step1
```

---

## Step 3: UI Updates (2-3 hours)

### Pre-Step Checkpoint:
- [ ] Step 2 completed successfully
- [ ] Read STEP3_UI_UPDATES.md completely
- [ ] Can load multiple profiles

### Implementation Tasks:
- [ ] Add `shareToProfile` map to MainWindow struct
- [ ] Initialize `shareToProfile` in NewMainWindow()
- [ ] Create custom share widget with rendering:
  - [ ] Profile headers (bold, 📁 icon)
  - [ ] Indented shares
- [ ] Update share selection handler:
  - [ ] Skip profile headers
  - [ ] Lookup profile via `shareToProfile` map
  - [ ] Set `t.selected.profileID`
- [ ] Add `refreshShareList()` function
- [ ] Update `loadSharesForProfile()` to call `refreshShareList()`
- [ ] Update `switchToProfile()` to call `refreshShareList()`
- [ ] Update toolbar with profile management buttons

### Testing:
- [ ] Share list shows grouped format
- [ ] Profile headers are bold and non-selectable
- [ ] Shares are indented and selectable
- [ ] Clicking share loads correct schemas
- [ ] Multiple profiles display correctly
- [ ] Toolbar buttons present

### Post-Step Actions:
- [ ] **Commit changes:**
  ```bash
  git add windows/mainWindow.go
  git commit -m "Step 3: Update UI with grouped share list"
  git tag checkpoint/step3
  ```

### Rollback Point:
```bash
# If Step 3 fails:
git reset --hard checkpoint/step2
```

---

## Step 4: Logic Integration (2-3 hours)

### Pre-Step Checkpoint:
- [ ] Step 3 completed successfully
- [ ] Read STEP4_LOGIC_INTEGRATION.md completely
- [ ] Share list displays correctly

### Implementation Tasks:
- [ ] Verify share selection handler (already done in Step 3)
- [ ] Update schema selection handler:
  - [ ] Validate `t.selected.profileID`
  - [ ] Clear dependent data only
- [ ] Update table selection handler:
  - [ ] Validate complete selection path
  - [ ] Get profile via `t.selected.profileID`
  - [ ] Pass profile content to dataBrowser
- [ ] Update table context menu handler:
  - [ ] Get profile via `t.selected.profileID`
  - [ ] Use profile content for client creation
- [ ] Update `ScanTree()` function:
  - [ ] Check for `t.selected.profileID`
  - [ ] Get profile from ProfileManager
  - [ ] Use profile's cached client
- [ ] Add cache helper functions (optional):
  - [ ] getCachedSchemas()
  - [ ] setCachedSchemas()

### Testing:
- [ ] Complete selection flow works (share → schema → table)
- [ ] Data loads from correct profile
- [ ] Multiple profiles don't interfere
- [ ] Context menu "Load with Options" works
- [ ] Client connections reused
- [ ] Error handling works

### Post-Step Actions:
- [ ] **Commit changes:**
  ```bash
  git add windows/mainWindow.go
  git commit -m "Step 4: Integrate profile logic with selection handlers"
  git tag checkpoint/step4
  ```

### Rollback Point:
```bash
# If Step 4 fails:
git reset --hard checkpoint/step3
```

---

## Step 5: Profile Management Dialog (2-3 hours)

### Pre-Step Checkpoint:
- [ ] Step 4 completed successfully
- [ ] Read STEP5_PROFILE_MANAGEMENT_DIALOG.md completely
- [ ] Data loads correctly from any profile

### Implementation Tasks:
- [ ] Create `windows/profileManagerDialog.go`
- [ ] Add ProfileManagerDialog struct
- [ ] Implement `createDialog()`
- [ ] Implement `updateDetailsView()`
- [ ] Implement profile operations:
  - [ ] setActiveProfile()
  - [ ] renameProfile()
  - [ ] reloadProfile()
  - [ ] closeProfile()
  - [ ] showProfileDetails()
- [ ] Implement bulk operations:
  - [ ] reloadAllProfiles()
  - [ ] closeAllProfiles()
- [ ] Implement `refreshDialog()`
- [ ] Add helper function `join()`
- [ ] Update `showProfileManagerDialog()` in mainWindow.go

### Testing:
- [ ] Dialog displays all profiles
- [ ] Profile selection updates details panel
- [ ] Set Active works (updates IsActive flag)
- [ ] Rename works and updates display
- [ ] Reload works and updates shares
- [ ] Close profile works
- [ ] View Details shows full info
- [ ] Reload All works
- [ ] Close All works and clears UI

### Post-Step Actions:
- [ ] **Commit changes:**
  ```bash
  git add windows/profileManagerDialog.go windows/mainWindow.go
  git commit -m "Step 5: Add comprehensive profile management dialog"
  git tag checkpoint/step5
  ```

### Rollback Point:
```bash
# If Step 5 fails:
git reset --hard checkpoint/step4
rm -f windows/profileManagerDialog.go
```

---

## Step 6: Polish & Persistence (2-3 hours)

### Pre-Step Checkpoint:
- [ ] Step 5 completed successfully
- [ ] Read STEP6_POLISH_PERSISTENCE.md completely
- [ ] Profile management dialog works

### Implementation Tasks:

#### Part A: Persistence
- [ ] Add SavedProfile struct to profileManager.go
- [ ] Add `SaveToPreferences()` to ProfileManager
- [ ] Add `LoadFromPreferences()` to ProfileManager
- [ ] Add `AddProfileWithID()` to ProfileManager
- [ ] Add `restoreProfiles()` to mainWindow.go
- [ ] Add `saveProfiles()` to mainWindow.go
- [ ] Call restore in NewMainWindow()
- [ ] Set up auto-save on window close
- [ ] Add save calls after profile operations

#### Part B: Validation
- [ ] Add `ValidateProfile()` to profileManager.go
- [ ] Use validation in OpenProfile()

#### Part C: Error Handling
- [ ] Add `showProfileError()` to mainWindow.go
- [ ] Replace generic error dialogs

#### Part D: Visual Improvements
- [ ] Add tooltips to toolbar buttons
- [ ] Add loading notifications in initializeAndLoadProfile()
- [ ] Add keyboard shortcuts (Ctrl+P, Ctrl+M, F5)
- [ ] Update `refreshShareList()` with status indicators

#### Part E: Health Check (Optional)
- [ ] Add `CheckProfileHealth()` to ProfileManager
- [ ] Add background health monitoring (optional)

### Testing:
- [ ] Profiles persist across app restarts
- [ ] Invalid profiles rejected with clear errors
- [ ] Error messages helpful and specific
- [ ] Notifications appear for loading/errors
- [ ] Keyboard shortcuts work
- [ ] Share list shows error indicators
- [ ] Auto-save works on close

### Post-Step Actions:
- [ ] **Commit changes:**
  ```bash
  git add windows/profileManager.go windows/mainWindow.go
  git commit -m "Step 6: Add polish and persistence features"
  git tag checkpoint/step6
  ```

### Rollback Point:
```bash
# If Step 6 fails:
git reset --hard checkpoint/step5
```

---

## Step 7: Testing (2-3 hours)

### Pre-Step Checkpoint:
- [ ] Step 6 completed successfully
- [ ] Read STEP7_TESTING.md completely
- [ ] All features implemented

### Testing Categories:

#### Basic Profile Operations:
- [ ] Test 1.1: Load single profile
- [ ] Test 1.2: Load multiple profiles (add to existing)
- [ ] Test 1.3: Replace all profiles
- [ ] Test 1.4: Close single profile
- [ ] Test 1.5: Close all profiles

#### Multi-Profile Scenarios:
- [ ] Test 2.1: Work with multiple profiles simultaneously
- [ ] Test 2.2: Switch between profiles via share selection
- [ ] Test 2.3: Data isolation between profiles
- [ ] Test 2.4: Profile with same share names

#### UI/UX Validation:
- [ ] Test 3.1: Grouped share list display
- [ ] Test 3.2: Profile headers non-selectable
- [ ] Test 3.3: Profile Management Dialog
- [ ] Test 3.4: Error indicators
- [ ] Test 3.5: Status bar updates

#### Data Integrity:
- [ ] Test 4.1: Correct data loads from correct profile
- [ ] Test 4.2: Schema/table navigation
- [ ] Test 4.3: Data browser shows correct data
- [ ] Test 4.4: Export functions work

#### Error Handling:
- [ ] Test 5.1: Invalid profile file
- [ ] Test 5.2: Network errors
- [ ] Test 5.3: Missing profile file
- [ ] Test 5.4: Corrupted profile data

#### Performance:
- [ ] Test 6.1: Load time with 5+ profiles
- [ ] Test 6.2: Memory usage with multiple profiles
- [ ] Test 6.3: Client connection reuse
- [ ] Test 6.4: UI responsiveness

#### Persistence:
- [ ] Test 7.1: Profiles restore on restart
- [ ] Test 7.2: Active profile restored
- [ ] Test 7.3: Auto-save on operations
- [ ] Test 7.4: Handle moved profile files

### Bug Tracking:
```
Bug #1:
Description:
Steps to Reproduce:
Expected:
Actual:
Severity: [Critical/High/Medium/Low]
Status: [Open/Fixed/Won't Fix]

Bug #2:
...
```

### Post-Step Actions:
- [ ] **Document test results**
- [ ] **Fix any critical bugs**
- [ ] **Commit final changes:**
  ```bash
  git add .
  git commit -m "Step 7: Testing complete, bugs fixed"
  git tag checkpoint/step7-complete
  ```

### Final Rollback Point:
```bash
# If testing reveals major issues:
git reset --hard checkpoint/step6
# Or completely rollback:
git reset --hard backup/before_multi_profile_<date>
```

---

## Final Integration

### Pre-Merge Checklist:
- [ ] All 7 steps completed
- [ ] All tests passed
- [ ] No known critical bugs
- [ ] Documentation updated
- [ ] CLAUDE.md updated if needed
- [ ] Code reviewed (self or peer)
- [ ] Performance acceptable
- [ ] Memory usage reasonable

### Merge to Main:
```bash
# Ensure feature branch is up to date
git checkout feature/multi_profile
git fetch origin main
git merge origin/main
# Resolve any conflicts

# Run final tests
go build
./dsb  # Full manual test

# Merge to main
git checkout main
git merge feature/multi_profile --no-ff
git push origin main

# Tag release
git tag v2.0.0-multi-profile
git push origin v2.0.0-multi-profile

# Keep feature branch as backup
git branch backup/multi_profile_release
```

---

## Post-Implementation

### Cleanup:
- [ ] Archive old backup branches (after 30 days)
- [ ] Update README with new features
- [ ] Create user documentation
- [ ] Announce new feature to users

### Monitoring:
- [ ] Watch for user-reported issues
- [ ] Monitor memory usage in production
- [ ] Track performance metrics
- [ ] Gather user feedback

---

## Emergency Rollback After Deployment

If critical issues discovered in production:

```bash
# Immediate rollback
git checkout main
git revert <merge-commit> --mainline 1
git push origin main

# Or hard reset (if not shared)
git reset --hard <commit-before-merge>
git push origin main --force-with-lease

# Document incident
# Follow ROLLBACK_PLAN.md Type D
```

---

## Progress Tracking

Use this section to track overall progress:

**Start Date:** _______________
**Target Completion:** _______________

| Step | Status | Start Date | End Date | Time Spent | Issues |
|------|--------|------------|----------|------------|--------|
| 1. Foundation | ⬜ | | | | |
| 2. Profile Loading | ⬜ | | | | |
| 3. UI Updates | ⬜ | | | | |
| 4. Logic Integration | ⬜ | | | | |
| 5. Profile Dialog | ⬜ | | | | |
| 6. Polish | ⬜ | | | | |
| 7. Testing | ⬜ | | | | |
| Final Integration | ⬜ | | | | |

**Status Legend:** ⬜ Not Started | 🟡 In Progress | ✅ Complete | ❌ Failed

**Total Time Estimate:** 15-22 hours
**Actual Time:** _______________

---

## Notes and Lessons Learned

Use this section to document:
- Unexpected challenges
- Solutions to problems
- Things that worked well
- Things to improve next time

```
Date: _______________
Note: _______________

Date: _______________
Note: _______________
```

---

**Related Documents:**
- [README.md](README.md) - Overview and architecture
- [ROLLBACK_PLAN.md](ROLLBACK_PLAN.md) - Complete rollback instructions
- [ROLLBACK_QUICK_REFERENCE.md](ROLLBACK_QUICK_REFERENCE.md) - Quick rollback commands
- [STEP1_FOUNDATION.md](STEP1_FOUNDATION.md) through [STEP7_TESTING.md](STEP7_TESTING.md) - Detailed implementation steps

---

*Version: 1.0*
*Last Updated: 2025-10-17*
