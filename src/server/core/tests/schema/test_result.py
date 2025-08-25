from acontext_core.schema.result import Result, Code
from fastapi.responses import JSONResponse


def test_result_class():
    test_data = {"message": "pong"}
    suc = Result.resolve(test_data)
    d, eil = suc.unpack()
    assert d == test_data
    assert eil is None

    err = Result.reject(Code.BAD_REQUEST, "test")
    d, eil = err.unpack()
    assert d is None
    assert eil.status == Code.BAD_REQUEST
