gcp-path - Google Cloud Platform resource hierarchy
=========================================================
gcp-path is a utility to make it easier to work with gcloud from the command line on
the Google Cloud Platform resource hierarchy.

## ls folders
To  recursive list all folders in your organizations, type:
```
$ gcp-path ls
//xebia.com/sl
//xebia.com/sl/cloud
//xebia.com/sl/cloud/playgrounds
//xebia.com/sl/data
//xebia.com/sl/transformation
//xebia.com/sl/microsoft
//xebia.com/sl/software%20technology
//xebia.com/playgrounds/slash%2F-and-burn
...
```
The paths will be proper URL paths, so the names may contain slashes and spaces. These will be encoded.

## get Google Cloud Platform resource name by path
to get the resource name of a path, type:

```
$ gcp-path get-resource-name //xebia.com
organizations/2342342342334

$ gcp-path get-resource-name //xebia.com/sl/cloud/playgrounds
folders/134234534556
```
The name must be a proper path.
Some gcloud commands do not accept the name, but only the id. In that case, add the flag --id:

```shell
$ gcp-path get-resource-name --id //xebia.com
2342342342334
```

You can use it directly in a gcloud command, as shown below:
```
gcloud resource-manager folders list \
    --organization $(gcp-path get-resource-name --id //xebia.com)
```

## get path of a resource name
You can also get the path of a Google Cloud Platform resource manager resource, just type:

```
$ gcp-path get-path organizations/2342342342334
//xebia.com

$ gcp-path get-path folders/134234534556
//xebia.com/sl/cloud/playgrounds
```

# installation
To install the utility, type:

```
go install github.com/xebia/gcp-path
```

# CAVEATS
- The Google cloud resource manager folders API is dog slow listing folders.
