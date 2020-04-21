/*
Package protocol includes some basics for the Steam protocol.

It defines basic interfaces that are used throughout go-steam:

There is Message, which is extended by ClientMessage (sent after logging in) and abstracts over the
outgoing message types. Both interfaces are implemented by ClientProtoMessage and
ClientStructMessage. StructMessage is like ClientStructMessage, but it is used for sending messages
before logging in.

There is also the concept of a Packet: this is a type for incoming messages where only the header is
deserialized. It therefore only contains EMsg data, job information and the remaining data. Its
contents can then be read via the Read* methods which read data into a MessageBody - a type which is
Serializable and has an EMsg.

In addition, there are extra types for communication with the Game Coordinator (GC) included in the
gc sub-package. For outgoing messages the gc.Message interface is used which is implemented by
gc.ProtoMessage and gc.StructMessage. Incoming messages are of the gc.Packet type and are read like
regular Packets.

The actual messages and enums are in an external repository at
https://github.com/13k/go-steam-resources.
*/
package protocol
