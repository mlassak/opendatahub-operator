## Integration testing

A Jenkins pipeline for integration tests is set up as an additional testing layer for pull requests going into ODH `main` branch.
This test workflow is recommended to be used once the PR author feels that their changes are ready for full testing before merge.

To run this test pipeline on a PR, please follow the steps below:

1. when ready, label the PR as ready for integration tests
   - by commenting `/label run-integration-tests`
   - `openshift-ci` bot will add the label to your PR
2. once label is present, each following push into the PR branch will trigger the integration tests pipeline. More specifically, the following process will happen:
   1. `Build Catalog FBC and run Integration tests` GitHub Action will be run
      - this action builds the catalog image from the PR branch, and then pushes it to the image registry
      - in case of failure, please check GitHub Action console output for troubleshooting
   2. once that pre-requisite job succeeds, `github-actions` bot will comment `/test-integration`
   3. upon that event, integration tests pipeline is started in Jenkins