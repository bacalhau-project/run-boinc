# Run BOINC

This is a project to make it easier to run [BOINC](https://boinc.berkeley.edu/)
[workloads](https://boinc.berkeley.edu/projects.php) on [Bacalhau](https://www.bacalhau.org/).

# Building
1. [Install Just](https://github.com/casey/just)
2. Run `./justfile`
3. Binary is available at `./bin/run_boinc`

# Running
Before running a BOINC project on Bacalhau, you need:
* Sign up to a project and get your **weak** account key [^1]
* Determine what domains or IP addresses the project communicate with [^2]
* Make sure that the domains and IP addresses are allow-listed by Bacalhau

Example of running [latinsquares](https://boinc.multi-pool.info/latinsquares/) project:
`./bin/run-boinc --project-url https://boinc.multi-pool.info/latinsquares/ --weak-account-key <weakAccountKey> --domain 78.26.93.125 --domain boinc.berkeley.edu --domain boinc.multi-pool.info`

Example of running [einsteinathome.org](https://einsteinathome.org/):
`./bin/run-boinc --project-url https://einsteinathome.org/ --weak-account-key <weakAccountKey> --domain einsteinathome.org --domain scheduler.einsteinathome.org --domain einstein.phys.uwm.edu --domain einstein-dl.syr.edu --domain .aei.uni-hannover.de`

# Future tasks
* Spot when a task fails and do something about it

[^1] The non-weak account key is directly tied to your account whereas the weak account key is tied to your password,
so rotating the password will change the weak account key. There is no way to change your non-weak account key. Also,
the weak account key [does not allow someone to log into your account or change it in any way](https://boinc.berkeley.edu/wiki/Weak_account_key). Note that use of the weak-account key isn't enforced, but strongly recommended.

[^2] This may have to be done by running the Docker image locally a number of times and then finding out what domains
it will communicate with (`cat client_state.xml | grep http | cut -d'>' -f2 | cut -d/ -f3 | sort -u`)
