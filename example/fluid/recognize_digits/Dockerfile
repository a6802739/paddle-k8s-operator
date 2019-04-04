FROM paddlepaddle/paddle
RUN pip install -U kubernetes opencv-python && apt-get update -y && apt-get install -y iputils-ping libgtk2.0-dev
RUN python -m pip install pip==9.0.1
ADD paddle_k8s /usr/bin
ADD k8s_tools.py /root/
RUN chmod u+x /usr/bin/paddle_k8s
RUN mkdir -p /workspace
ADD recognize_digits.py /workspace
ADD common.py /workspace
CMD ["paddle_k8s"]
