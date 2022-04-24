# Uni-T UT61E+ USB Protocol Client

The [Uni-T UT61E+](https://www.uni-trend.com/meters/html/product/NewProducts/UT61%20161%20Series/UT61E+.html)
digital multimeter can be connected to a PC through the built-in CP2110 UART-USB bridge. This repository contains a
client library for its proprietary protocol.

## Using the library

### Dependencies

The library is using the cross-platform HIDAPI to communicate with the device. On a Debian based system installing the
following packages should be sufficient:
```bash
sudo apt install libhidapi-dev libudev-dev
```
The library is also available for other operating systems but this is untested.

## Example
```golang
dev, err := ut61ep.Open("")
if err != nil {
    log.Fatalf("open: %v", err)
}
message, err := dev.ReadMessage()
if err != nil {
    log.Fatalf("readMessage: %v", err)
}
log.Printf("%f %s", message.Value, message.Unit.String())
```

## Protocol Description

### Configuring CP2110
Before any communication with the multimeter can take place the CP2110 UART-USB bridge chip has to be configured by sending feature reports:

```
// Enable UART
0x41, 0x01

// Configure UART (9600 baud, parity=NONE, 8 Bit, hardware Flow Control disabled, stop bits short)
0x50, 0x00, 0x00, 0x25, 0x80, 0x00, 0x00, 0x03, 0x00, 0x00
```

### Requesting data from the device

Data can be requested by sending the following byte sequence:
```
0x06, 0xab, 0xcd, 0x03, 0x5e, 0x01, 0xd9
```

### Reply from the device
The device always sends two bytes at a time, the first of which is always a 1 and the second one containing the actual
data byte. The following message description ignores the first byte and shows only the actual data bytes.

The device responds with a 19 byte long message. The first two bytes are constant, the following byte contains the
message length and the remainder is the actual data.

```
01 AB CD 10 00 00 30 20 30 2e 30 38 35 30 00 01 30 30 30 03
-- ----- --    -- --    -----------------             --
 |   |    |     |  |            |                      |
 |   |    |     |  |            |                      |
 |   |    |     |  |            |                      |
 |   |    |     |  |            |                       ------ sign / sub-mode
 |   |    |     |  |             ----------------------------- value
 |   |    |     |   ------------------------------------------ range
 |   |    |      --------------------------------------------- mode
 |   |     --------------------------------------------------- length of the following message
 |    -------------------------------------------------------- constant preamble
  ------------------------------------------------------------ message type
```

For many of the bytes only the least significant 4-bits are relevant and the most significant 4-bits are 0x30.

## References
- This library was inspired by a [similar project](http://www.smartypies.com/projects/ut171a-data-reader-on-linux/) for a different Uni-T multimeter
- [CP2110 datasheet](https://www.silabs.com/documents/public/application-notes/an434-cp2110-4-interface-specification.pdf)