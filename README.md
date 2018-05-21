# windows-service-riemann
Simple agent to send status of Windows services to Riemann


## Usage

Edit `config.yaml` with the hostname of your riemann instance.  Add the Window service name and desired state to the whitelist.  An `ok` or `critical` will be sent to riemann based on if the queried state matches the desired state.
