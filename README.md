### Functional Requirements

1. Fetch IP address from multiple sources (can be chosen by the user);
2. Scripts can be either Python or Bash.
3. App should expose an API for the changes and checks;
4. App should have extensive logging;
5. App should be able to notify users by email;
6. App configuration should be done via a YAML file.

### User Requirements

1. User can chose ways to be notified if a change happens;
2. User can define the delta between checks;
3. User can specify a script to run if the address has, or not changed;
4. User can choose to check IPv6, IPv4 or both ips.
 
### Configuration

1. url only needs v4 or v6, or both to be specified
2. url must be a valid address
3. response_type is mandatory
4. response_type can be json | text 
5. field is mandatory when response_type is json
6. field must be present in the response from the api and represents the ip field
7. default source is always the first, the others are used as fallbac1. 
8. timeout is specified in seconds 