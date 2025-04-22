# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [isochrones.proto](#isochrones-proto)
    - [Isochrone](#horizon-Isochrone)
    - [IsochronesRequest](#horizon-IsochronesRequest)
    - [IsochronesResponse](#horizon-IsochronesResponse)
  
- [map_match.proto](#map_match-proto)
    - [MapMatchRequest](#horizon-MapMatchRequest)
    - [MapMatchResponse](#horizon-MapMatchResponse)
  
- [point.proto](#point-proto)
    - [GeoPoint](#horizon-GeoPoint)
  
- [service.proto](#service-proto)
    - [Service](#horizon-Service)
  
- [shortest_path.proto](#shortest_path-proto)
    - [SPRequest](#horizon-SPRequest)
    - [SPResponse](#horizon-SPResponse)
  
- [Scalar Value Types](#scalar-value-types)



<a name="isochrones-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## isochrones.proto



<a name="horizon-Isochrone"></a>

### Isochrone



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int64](#int64) |  | Isochrone ID |
| cost | [double](#double) |  | Travel cost to the target vertex |
| vertex_id | [int64](#int64) |  | Vertex ID in the graph |
| point | [GeoPoint](#horizon-GeoPoint) |  | Longitude, Latitude |






<a name="horizon-IsochronesRequest"></a>

### IsochronesRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| max_cost | [double](#double) | optional | Max cost restrictions for single isochrone. Should be in range [0,&#43;Inf]. Minumim is 0. Example: 2100.0 |
| max_nearest_radius | [double](#double) | optional | Max radius of search for nearest vertex (Optional, default is 25.0, should be in range [0,&#43;Inf]) Example: 25.0 |
| lon | [double](#double) |  | Longitude Example: 37.601249363208915 |
| lat | [double](#double) |  | Latitude Example: 55.745374309126895 |






<a name="horizon-IsochronesResponse"></a>

### IsochronesResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| isochrones | [Isochrone](#horizon-Isochrone) | repeated | List of isochrones |
| warnings | [string](#string) | repeated | List of warnings |





 

 

 

 



<a name="map_match-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## map_match.proto



<a name="horizon-MapMatchRequest"></a>

### MapMatchRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| req | [int64](#int64) |  | @todo |






<a name="horizon-MapMatchResponse"></a>

### MapMatchResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| res | [int64](#int64) |  | @todo |





 

 

 

 



<a name="point-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## point.proto



<a name="horizon-GeoPoint"></a>

### GeoPoint



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| lon | [double](#double) |  | Longitude Example: 37.601249363208915 |
| lat | [double](#double) |  | Latitude Example: 55.745374309126895 |





 

 

 

 



<a name="service-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## service.proto


 

 

 


<a name="horizon-Service"></a>

### Service


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| RunMapMatch | [MapMatchRequest](#horizon-MapMatchRequest) | [MapMatchResponse](#horizon-MapMatchResponse) |  |
| GetSP | [SPRequest](#horizon-SPRequest) | [SPResponse](#horizon-SPResponse) |  |
| GetIsochrones | [IsochronesRequest](#horizon-IsochronesRequest) | [IsochronesResponse](#horizon-IsochronesResponse) |  |

 



<a name="shortest_path-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## shortest_path.proto



<a name="horizon-SPRequest"></a>

### SPRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| req | [int64](#int64) |  | @todo |






<a name="horizon-SPResponse"></a>

### SPResponse



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| res | [int64](#int64) |  | @todo |





 

 

 

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

