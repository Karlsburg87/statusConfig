# statusConfig

Purposed to auto generate config files for statusSentry VM run services that is designed to fetch a config file of shape saveSentry.Configuration.Configuration and collect updates and pings based upon it - updating the config used on every receipt of an http call to the correct endpoint.

By generating this on the fly multiple statusSentry VM services working in unison can be balanced by changing the configuration delivered to each VM depending on the global configuration load.

## http interface

|**Endpoint**|Method|Query Key / Payload|Description|
|-|-|-|-|
|`/`|GET|-|html interface for administrators of the configuration list|
|`/get`|GET|?`q`=${statusName}|gets Config from the configuration datastore using the Service Name value under `q`|
|`/add`|POST|`Config` object|commits the Config object to the configuration datastore|
|`/remove`|DELETE|`Config` object *(only Service Name needed)*|removes the given Config with matching Service Name from the configuration datastore|
|`/update`|PATCH|`Config` object|merges and updates the existing value with the payload value|
|`/configuration`|GET|-|fetches the full Configuration. Used by statusSentry instances to get the configs to run from|

## settings
|**envar**|Description|
|-|-|
|`STATUSSENTRY_INSTANCES`|comma seperated list of URL endpoints of statusSentry instances used to nudge the instance to call for configuration refetch|
|`PORT`|port to use for the main server|
|||