def has_close_elements(numbers: list[float], threshold: float) -> bool:
    if len(numbers) < 2:
        return False
    sorted_numbers = sorted(numbers)
    for i in range(1, len(sorted_numbers)):
        if sorted_numbers[i] - sorted_numbers[i-1] < threshold:
            return True
    return False