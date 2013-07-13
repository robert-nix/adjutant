# adjutant - github auto-deploy server

Creates and maintains children according to repo config in adjutant.json.  An
executed shell script may have git clone, pull, etc. for fetching the upstream
repo, and install/execution instructions for actual deployment.  This script is
triggered via a github post-receive hook POSTing to the configured url.

Child standard pipes aren't used, so children will need to log to their own
files.
