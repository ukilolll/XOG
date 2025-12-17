from fastapi import FastAPI
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles

app = FastAPI()

# Serve index.html at root
@app.get("/")
def read_root():
    return FileResponse("index.html")

# (Optional) Serve additional static files
app.mount("/static", StaticFiles(directory="static"), name="static")