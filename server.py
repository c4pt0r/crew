# Example JSONRPC server for crew remote page rendering
from jsonrpcserver import Success, method, serve, Result 
import json

# `Render` method is called by the crew server to render a page
# returning the HTML in string format
@method
def Render(url : str, params: dict) -> Result:
    return Success("<h1>Hello From Remote</h1>" + url + json.dumps(params))

serve('localhost', 5001)
