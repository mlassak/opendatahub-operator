<!--- 
Many thanks for submitting your Pull Request ❤️!

Please complete the following sections for a smooth review.
-->

## Description
<!--- Describe your changes in detail -->

<!--- Link your JIRA and related links here for reference. -->

## How Has This Been Tested?
<!--- Please describe in detail how you tested your changes. -->
<!--- Include details of your testing environment and the tests you ran to -->
<!--- see how your change affects other areas of the code, etc. -->

## Screenshot or short clip
<!--- If applicable, attach a screenshot or a short clip demonstrating the feature. -->

## Merge criteria
<!--- This PR will be merged by any repository approver when it meets all the points in the checklist -->
<!--- Go over all the following points, and put an `x` in all the boxes that apply. -->

- [ ] You have read the [contributors guide](https://github.com/opendatahub-io/opendatahub-operator/blob/incubation/CONTRIBUTING.md).
- [ ] Commit messages are meaningful - have a clear and concise summary and detailed explanation of what was changed and why.
- [ ] Pull Request contains a description of the solution, a link to the JIRA issue, and to any dependent or related Pull Request.
- [ ] Testing instructions have been added in the PR body (for PRs involving changes that are not immediately obvious).
- [ ] The developer has manually tested the changes and verified that the changes work
- [ ] The developer has run the integration test pipeline and verified that it passed successfully

### E2E test suite update requirement

When bringing new changes to the operator code, such changes are by default required to be accompanied by extending and/or updating the E2E test suite accordingly. A GitHub Action check is in place to enforce this requirement.

To opt-out of this requirement:
1. Inspect the opt-out guidelines rules in the sections below
  - to determine if the nature of the PR changes allows for skipping this requirement
2. Create opt-out justification PR comment
  - start the comment with 'E2E update requirement opt-out justification:', and provide a short summary of reasons for opting-out of this requirement
3. Edit PR description to check the checkbox below: 

- [ ] Skip requirement to update E2E test suite for this PR

#### Appropriate cases for E2E test suite update opt-out:

- Documentation-only changes (README, comments, etc.)
- Unit test additions/modifications without functional changes
- Code style/formatting changes
- Dependency version updates without functional impact
- Build system changes that don't affect runtime behavior
- Non-functional refactoring with existing test coverage

#### NOT Appropriate cases for E2E test suite update opt-out:

- New feature implementations
- Bug fixes affecting user-facing functionality
- API changes or modifications
- Configuration changes affecting deployment
- Changes to controllers, operators, or core logic
- Cross-component integration modifications
- Changes affecting user workflows or UI
