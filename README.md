# ip-monitor
I wanted a way to check my IPv4, IPv6 and JioFiber WAN-IP for troubleshooting purposes and I thought why not just automate it?

Whenever there is a change in your IPv4, IPv6 and Jio WAN-IP, a telegram chat message will be sent out.

# Steps
* Add a .env file with your `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`, Jio WAN `username` and `password`, `dpipv6` (0 or 1).
* Run this command in the project root `go run .`

You can find a lot of online tutorials about creating a telegram bot, so I am not including that here.

## But why?
* While working from home, my corporate VPN kept disconnecting randomly, which was obviously very annoying. What’s worse, it was happening only to me, not to anyone else on my team.
* While playing Overwatch 2, the server would also disconnect at random.
* It occurred to me that this might be due to public IP rotation, but could it really be that frequent? I noted down my IP address before and after a disconnect and voilà! It had changed!
* So, I built this app to monitor my public IP and WAN IP every 2 minutes, and it confirmed that this was the problem.
* WAN-IP didn't change as frequently though, and I am not sure if that will be a problem, since whenever it changed (mostly during late night—early morning, I was not connected to the VPN (it changed very rarely).
* …and of course, Jio customer support had no idea what I was talking about, so nothing changed.
