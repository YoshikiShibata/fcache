# Function-Result cache

This is a simple function-result cache: a supplied function will be executed periodically and its result will be cached for later query.

When `Get` is called, then the supplied function will be invoked to get the lastest result. But when the function will note return within a specified timeout, then any previously cached result will be returned.
