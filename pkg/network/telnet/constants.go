package telnet

// Telnet commands
const (
    IAC  = 255 // Interpret as command
    DONT = 254 // You are not to use option
    DO   = 253 // Please use option
    WONT = 252 // I won't use option
    WILL = 251 // I will use option
    SB   = 250 // Subnegotiation Begin
    SE   = 240 // Subnegotiation End
    NOP  = 241 // No operation
    DM   = 242 // Data Mark
    BRK  = 243 // Break
    IP   = 244 // Interrupt Process
    AO   = 245 // Abort Output
    AYT  = 246 // Are You There
    EC   = 247 // Erase Character
    EL   = 248 // Erase Line
    GA   = 249 // Go Ahead
    EOR  = 239 // End of Record
)

// Telnet options
const (
    BINARY = 0    // Binary Transmission
    ECHO   = 1    // Echo
    SGA    = 3    // Suppress Go Ahead
    STATUS = 5    // Status
    TM     = 6    // Timing Mark
    TTYPE  = 24   // Terminal Type
    NAWS   = 31   // Negotiate About Window Size
    TSPEED = 32   // Terminal Speed
    LFLOW  = 33   // Remote Flow Control
    LINEMD = 34   // Line Mode
    XDISPL = 35   // X Display Location
    ENVIRON= 36   // Environment Option
    NENVIR = 39   // New Environment Option
    MSDP   = 69   // MUD Server Data Protocol
    MSSP   = 70   // MUD Server Status Protocol
    MCCP2  = 86   // MUD Client Compression Protocol v2
    MSP    = 90   // MUD Sound Protocol
    MXP    = 91   // MUD eXtension Protocol
    GMCP   = 201  // Generic MUD Communication Protocol
)
