# VFD Clock

### Bedside clock with extended info

I'm a sucker for [old display technologies](https://github.com/rschlaikjer/ntp-nixie),
so on my arduino binge I built a bedside clock capable of displaying generic
information, with long line scrolling and whatnot. The design is fully PoE, so
the finished clock just needs one cable.

**In this repo:**
- Arduino code for the clock logic, handles fetching data from the server and VFD control.
- Eagle schematics for the clock control PCB
- Housing DXF for laser cut housing. The design is intended to be cut on a 6.7mm material.
- Server code manages actual information fetching (checking my unread mail, getting the price of bitcoin)

## Partlist
- Atmega 328P (along with 16MHz crystal, 2x 22pf caps)
- Silvertel A9700 PoE module
- ENC28J60 ethernet PHY
- RB1-125BHQ1A RJ45 jack with magnetics
- LM1117 3.3V regulator (for ENC28J60)
- VFD with serial interface (Here I'm using a Noritake CU20045SCPB-T31A)

## Pretty pictures

![Front](/Pics/encased-front.jpg?raw=true "Front view")
![Side](/Pics/encased-side.jpg?raw=true "Side view")
![Proto](/Pics/prototype.jpg?raw=true "Initial testing")
