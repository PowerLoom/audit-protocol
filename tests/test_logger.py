import logging
import sys

formatter = logging.Formatter(u"%(levelname)-8s %(name)-4s %(asctime)s,%(msecs)d %(module)s-%(funcName)s: %(message)s")

stdout_handler = logging.StreamHandler(sys.stdout)
stdout_handler.setLevel(logging.DEBUG)
# stdout_handler.setFormatter(formatter)

test_logger = logging.getLogger(__name__)
test_logger.setLevel(logging.DEBUG)
test_logger.addHandler(stdout_handler)
