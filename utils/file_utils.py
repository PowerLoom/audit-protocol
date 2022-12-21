import os

def read_text_file(file_path: str, logger):
    """Read given json file and return its content as a dictionary."""
    try:
        file_obj = open(file_path, 'r', encoding='utf-8')
    except FileNotFoundError:
        return None
    except Exception as exc:
        logger.warning(f"Unable to open the {file_path} file")
        logger.error(exc, exc_info=True)
        return None
    else:
        return file_obj.read()


def write_bytes_to_file(directory:str, file_name: str, data, logger):
    try:
        file_path = directory + file_name
        if not os.path.exists(directory):
            os.makedirs(directory)
        file_obj = open(file_path, 'wb')
    except Exception as exc:
        logger.error(f"Unable to open the {file_path} file")
        raise exc
    else:
        bytes_written = file_obj.write(data)
        logger.debug("Wrote %s bytes to file %s",bytes_written, file_path)
        file_obj.close()