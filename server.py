from jsonrpcserver import Success, method, serve, InvalidParams, Result, Error
import re

@method(name="Render")
def Render(url, params) -> Result:    
    return Success("Hello" + url)

serve('localhost', 5001)
