#!/bin/bash
ipfw add 100 fwd 127.0.0.1,1080  tcp from any to me 80
ipfw add 100 fwd 127.0.0.1,10443 tcp from any to me 443
