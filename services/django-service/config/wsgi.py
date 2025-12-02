# config/wsgi.py
import os
from django.core.wsgi import get_wsgi_application

os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'config.settings')

# Setup tracing before loading application
from config.tracing import setup_tracing
setup_tracing()

application = get_wsgi_application()