# SIDER

SIDER is my understanding and implementation of REDIS.
### Current Progress
- Synchronous TCP Server — basic TCP server implemented and working.

- RESP2 Decoder — implemented support for:
    - Simple Strings
    - Errors
    - Integers
    - Bulk Strings
    - Arrays
    - Null Bulk Strings / Arrays
    - RESP Tests — test suite covering the implemented RESP types and invalid/incomplete inputs.
    Tests are written by AI.
- RESP2 Encoder - Implemented encoder for the above data types.
- Commands : 
    - PING
    - SET
    - GET
    - TTL
    - DEL
    - EXPIRE

- TODO : 
    - Persistence 
    - Key eviction strategy

