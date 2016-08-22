
https://www.skyscanner.net/transport/flights/lond/tls/161101/161103/

http://partners.api.skyscanner.net/apiservices/browseroutes/v1.0/GB/GBP/en-GB/LON/anywhere/20160819/20160821?apiKey=KEY
http://partners.api.skyscanner.net/apiservices/browseroutes/v1.0/GB/GBP/en-GB/LON/anywhere/20161101/20161103?apiKey=ar926739961631567929873917891697




TODO:
- check if bolt db multithreaded
- clean up main
  - check naming
  - move file around (request, reply, utils)
- try to have a generic parsing function
- rename
- How to fix the file issue (defer not being dealed with)
- add live query to cache
  - how to reuse local
- Move to skylib
- change githup name and delete (no more key)
- find git ignore file and add db and key
- add argument object

Naming:
  - Request
  - Reply
  - Do not use Query
  - normalize inbound/outbound vs departure / return
  - normalize key/api key

Header of live request; 
map[Content-Type:[application/json] Date:[Tue, 16 Aug 2016 19:54:45 GMT] Location:[http://partners.api.skyscanner.net/apiservices/pricing/uk2/v1.0/7973f72c4c37493c9d4fddf626a2efc8_ecilpojl_96FF83C678D7E7C1F630A16E8F3068D6] Content-Length:[2] Cache-Control:[private]]

