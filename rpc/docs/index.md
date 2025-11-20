# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [isochrones.proto](#isochrones-proto)
    - [Isochrone](#horizon-Isochrone)
    - [IsochronesRequest](#horizon-IsochronesRequest)
    - [IsochronesResponse](#horizon-IsochronesResponse)
  
- [map_match.proto](#map_match-proto)
    - [GPSToMapMatch](#horizon-GPSToMapMatch)
    - [IntermediateEdge](#horizon-IntermediateEdge)
    - [MapMatchRequest](#horizon-MapMatchRequest)
    - [MapMatchResponse](#horizon-MapMatchResponse)
    - [ObservationEdge](#horizon-ObservationEdge)
    - [SubMatch](#horizon-SubMatch)
  
- [point.proto](#point-proto)
    - [GeoPoint](#horizon-GeoPoint)
  
- [service.proto](#service-proto)
    - [Service](#horizon-Service)
  
- [shortest_path.proto](#shortest_path-proto)
    - [EdgeInfo](#horizon-EdgeInfo)
    - [SPRequest](#horizon-SPRequest)
    - [SPResponse](#horizon-SPResponse)
  
- [Scalar Value Types](#scalar-value-types)



<a name="isochrones-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## isochrones.proto



<a name="horizon-Isochrone"></a>

### Isochrone
Single isochrone information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [int64](#int64) |  | Isochrone ID |
| cost | [double](#double) |  | Travel cost to the target vertex |
| vertex_id | [int64](#int64) |  | Vertex ID in the graph |
| point | [GeoPoint](#horizon-GeoPoint) |  | Longitude, Latitude |






<a name="horizon-IsochronesRequest"></a>

### IsochronesRequest
User&#39;s request for isochrones


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| max_cost | [double](#double) | optional | Max cost restrictions for single isochrone. Should be in range [0,&#43;Inf]. Minumim is 0. Example: 2100.0 |
| max_nearest_radius | [double](#double) | optional | Max radius of search for nearest vertex (Optional, default is 25.0, should be in range [0,&#43;Inf]) Example: 25.0 |
| lon | [double](#double) |  | Longitude Example: 37.601249363208915 |
| lat | [double](#double) |  | Latitude Example: 55.745374309126895 |






<a name="horizon-IsochronesResponse"></a>

### IsochronesResponse
Server&#39;s response for isochrones request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| isochrones | [Isochrone](#horizon-Isochrone) | repeated | List of isochrones |
| warnings | [string](#string) | repeated | List of warnings |





 

 

 

 



<a name="map_match-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## map_match.proto



<a name="horizon-GPSToMapMatch"></a>

### GPSToMapMatch
Representation of GPS data


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tm | [string](#string) |  | Timestamp. Field would be ignored for request on &#39;/shortest&#39; service. Example: 2020-03-11T00:00:00 |
| lon | [double](#double) |  | Longitude Example: 37.601249363208915 |
| lat | [double](#double) |  | Latitude Example: 55.745374309126895 |
| accuracy | [double](#double) | optional | GPS measurement accuracy in meters (optional, &lt;=0 or null means use default sigma) Example: 5.0 |






<a name="horizon-IntermediateEdge"></a>

### IntermediateEdge
Edge which is not matched to any observation but helps to form whole travel path


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| geom | [GeoPoint](#horizon-GeoPoint) | repeated | Edge geometry as line feature |
| weight | [double](#double) |  | Travel cost Example: 2.0 |
| id | [int64](#int64) |  | Edge identifier Example: 4278 |






<a name="horizon-MapMatchRequest"></a>

### MapMatchRequest
User&#39;s request for map matching


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| max_states | [int32](#int32) | optional | Max number of states for single GPS point (in range [1, 10], default is 5). Field would be ignored for request on &#39;/shortest&#39; service. Example: 5 |
| state_radius | [double](#double) | optional | Max radius of search for potential candidates (in range [7, 50], default is 25.0) Example: 7.0 |
| gps | [GPSToMapMatch](#horizon-GPSToMapMatch) | repeated | Set of GPS data |






<a name="horizon-MapMatchResponse"></a>

### MapMatchResponse
Server&#39;s response for map matching request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| sub_matches | [SubMatch](#horizon-SubMatch) | repeated | Array of sub-matches (segments split when route cannot be computed between consecutive points) |
| warnings | [string](#string) | repeated | List of warnings |






<a name="horizon-ObservationEdge"></a>

### ObservationEdge
Relation between observation and matched edge


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| obs_idx | [int32](#int32) |  | Index of an observation. Index correspondes to index in incoming request. If some indices are not presented then it means that they have been trimmed Example: 0 |
| edge_id | [int64](#int64) |  | Matched edge identifier Example: 3149 |
| vertex_id | [int64](#int64) |  | Matched vertex identifier Example: 44014 |
| matched_edge | [GeoPoint](#horizon-GeoPoint) | repeated | Corresponding matched edge as line feature |
| matched_edge_cut | [GeoPoint](#horizon-GeoPoint) | repeated | Cut for excess part of the matched edge. Will be null for every observation except the first and the last. Could be null for first/last edge when projection point corresponds to source/target vertices of the edge |
| matched_vertex | [GeoPoint](#horizon-GeoPoint) |  | Corresponding matched vertex as point feature |
| projected_point | [GeoPoint](#horizon-GeoPoint) |  | Corresponding projection on the edge as point feature |
| next_edges | [IntermediateEdge](#horizon-IntermediateEdge) | repeated | Set of leading edges up to next observation (so these edges is not matched to any observation explicitly). Could be an empty array if observations are very close to each other or if it just last observation |






<a name="horizon-SubMatch"></a>

### SubMatch
A single continuous matched segment


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| observations | [ObservationEdge](#horizon-ObservationEdge) | repeated | Set of matched edges for observations in this segment |
| probability | [double](#double) |  | Probability from Viterbi algorithm for this segment Example: -86.578520 |





 

 

 

 



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



<a name="horizon-EdgeInfo"></a>

### EdgeInfo
Edge information


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| edge_id | [int64](#int64) |  |  |
| weight | [double](#double) |  | Travel cost to the target vertex |
| geom | [GeoPoint](#horizon-GeoPoint) | repeated | Line |






<a name="horizon-SPRequest"></a>

### SPRequest
User&#39;s request for finding shortest path


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| state_radius | [double](#double) | optional | Max radius of search for potential candidates (in range [7, 50], default is 25.0) Example: 10.0 |
| gps | [GeoPoint](#horizon-GeoPoint) | repeated | Set of GPS data |






<a name="horizon-SPResponse"></a>

### SPResponse
Server&#39;s response for shortest path request


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data | [EdgeInfo](#horizon-EdgeInfo) | repeated | List of edges in a path |
| warnings | [string](#string) | repeated | List of warnings |





 

 

 

 



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

