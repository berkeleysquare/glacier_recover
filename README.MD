#Glacier Recover

##Credentials
Credentials are stored in an AWS profile, like the AWS CLI. (If you omit the --profile <name> parameter you get the default.)

$ aws configure --profile myvail
(It will ask for the following, one line at a time. If configured, you get the [****1234] style
and you can just hit return to keep the exiting value)
```AWS Access Key ID [****************4564]:
AWS Secret Access Key [****************sw7e]:
Default region name [us-west-2]:
Default output format [json]:
```
Now you can just use the --profile myvail and not have to enter. But do note that if you do not specify an endpoint, it will perform the operation on AWS us-west-2 and not Vail.

##Installation
This is a self-contained application. For Windows, just unzip it into a directory. 
Then navigate to that dir in a command window.
Examples are run out of GitBash; Windows CMD will not require the ./

Test by displaying help:
```
$ ./glacier_recover --help
$ ./glacier_recover --command list_commands
```

##Cleaning missing Vail packs
### Export bucket inventory (optional)
The inventory command writes a bucket's object list to a .csv file (or stdout).
```
johnk@JK-P7530-LT MINGW64 /c/glacier_recover
$ ./glacier_recover.exe --command inventory  --endpoint https://10.85.41.101 --out jk-rio-inventory.csv --bucket jk-rio --profile myvail --no-verify-ssl
Ready
```
### Test run
The test_byte_restore command without the --delete-on-fail flag writes to a .csv file (or stdout):

- The object name
- whether it can be restored 
- if it was deleted (blank/no if --delete-on-fail is not set)
- any errors locating the key
- any errors deleting the object  
```
johnk@JK-P7530-LT MINGW64 /c/glacier_recover
$ ./glacier_recover.exe --command test_byte_restore  --endpoint https://10.85.41.101 --out jk-ps-44-test.csv --bucket jk-ps-44 --profile myvail --no-verify-ssl
Ready
```
### Clean (delete objects which can not be restored)
The test_byte_restore command with the --delete-on-fail flag writes to a .csv file (or stdout):

- The object name
- whether it can be restored 
- if it was deleted ("Deleted" if true, blank if not, "GLACIER" if ignored because Glacier class)
- any errors locating the key
- any errors deleting the object  
```
johnk@JK-P7530-LT MINGW64 /c/glacier_recover
$ ./glacier_recover.exe --command test_byte_restore  --endpoint https://10.85.41.101 --out jk-ps-44-clean.csv --bucket jk-ps-44 --delete-on-fail --profile myvail --no-verify-ssl
Ready
```
