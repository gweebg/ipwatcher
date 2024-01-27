### Functional Requirements

1. Fetch IP address from multiple sources (can be chosen by the user);
2. Scripts can be either Python or Bash.
3. App should expose an API for the changes and checks;
4. App should have extensive logging;
5. App should be able to notify users by email;
6. App configuration should be done via a YAML file;
7. If watcher is configured for both v4 and v6, it should spawn two processes;
8. 

### User Requirements

1. User can choose ways to be notified if a change happens;
2. User can define the delta between checks;
3. User can specify a script to run if the address has, or not changed;
4. User can choose to check IPv6, IPv4 or both ips.
