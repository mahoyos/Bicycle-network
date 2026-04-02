from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
import logging
from .database import create_tables
from . import models
from .routers import events

logging.basicConfig(level=logging.ERROR)

create_tables()

app = FastAPI(title="Events Microservice",
              description="A FastAPI microservice for managing events", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(events.router)


@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    logging.error(f"Unexpected error: {exc}")
    return JSONResponse(
        status_code=500,
        content={"message": "An unexpected error occurred. Please try again later."}
    )


@app.get("/")
def read_root():
    return {"message": "Welcome to the Events Microservice"}
