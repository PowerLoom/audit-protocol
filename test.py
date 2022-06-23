from tenacity import stop_after_attempt, retry_if_exception_type, retry
import sys

def is_false(value):
    """Return True if value is False"""
    return value is False

class IPFSOpException(Exception):
    pass

def return_last_value(retry_state):
        """return the result of the last call attempt"""
        print("inside return_last_value")
        try:
            result = retry_state.outcome.result()
        except Exception as err:
            print(f"there was exception in ipfs: {err} | {retry_state}")
            result = "status code 500"

        return result

# will return False after trying 3 times to get a different result
@retry(
    stop=stop_after_attempt(4),
    retry=retry_if_exception_type(IPFSOpException),
    retry_error_callback=return_last_value)
def eventually_return_false():
    print(f"here is ###@@@")
    raise IPFSOpException
    return "status code 200"


if __name__ == "__main__":
    try: 
        result = eventually_return_false()
        print(f"result: {result}")
    except Exception as err:
        pass