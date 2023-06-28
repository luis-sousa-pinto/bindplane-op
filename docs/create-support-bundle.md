# Create Support Bundle Script

The script located at [scripts/support/create-support-bundle.sh](../scripts/support/create-support-bundle.sh). It produce output in the directory it's run from, and collects the following information:

1.  BindPlane logs
2.  BindPlane configuration
3.  With `--agent` flag, the agent configuration and agent information
4.  System information

It will package the information into a tar.gz file named `bindplane_support_bundle_YYYYMMDD_HHMMSS.tar.gz`, where YYYY is the year, MM month, DD, day, HH hour, MM minute, SS second. It must be run with `sudo` or as root.
