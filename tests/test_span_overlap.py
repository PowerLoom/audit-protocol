from utils import retrieval_utils

span_a = (10, 20)
span_b = (5, 15)

intersection = retrieval_utils.check_intersection(span_a, span_b)
print(intersection)