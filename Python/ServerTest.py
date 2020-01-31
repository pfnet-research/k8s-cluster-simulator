from concurrent import futures
import logging

import grpc

import k8s_sim_pb2
import k8s_sim_pb2_grpc

class simRPCServicer(k8s_sim_pb2_grpc.simRPCServicer):
    def RecordMetrics(self, request, context):
        metrics = request
        clock = metrics.clock.clock_Key
        node = metrics.node.nodes_metrics_key
        pods = metrics.pods.pods_metrics_key
        queue = metrics.queue.queue_metrics_key

        print(clock)
        print(node)
        print(pods)
        print(queue)

        return 1

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    k8s_sim_pb2_grpc.add_simRPCServicer_to_server(
        simRPCServicer(), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    logging.basicConfig()
    serve()
