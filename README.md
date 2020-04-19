# PingCLI
PingCLI project for Cloudflare.

Ping CLI application for MacOS or Linux using goLang. The CLI app accepts a hostname or an IP address as its argument, then send ICMP "echo requests" in a loop to the target while receiving "echo reply" messages.

##### It reports loss and RTT times for each sent message.

 It also gives users features like:
 1. Set limit: Sets the limit to maximum successful requests. (Use -l="Limit")
 2. Set interval: Allows user to set how often a packet is sent. (Use -i="interval")
 4. RTT limit: In case this limit is surpassed report the user. (Use -rtt="limit")
 3. Display stats: Displays various statistics about packets and RTT (Use -stats=true)

