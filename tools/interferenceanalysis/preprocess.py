"""
Indicator data preprocessing module.

The low-level metrics do not correspond to the high-level SLAs,
and the indicators are corresponding and merged according to
the method of time stamp proximity search..
"""

import time
import sys
import pandas as pd
import numpy as np

class Preprocess(object):
    def __init__(self, metrics, qos, output):
        self.metrics = metrics
        self.qos = qos
        self.output = output

    def execute(self) -> None:
        """
        Execute preprocessing
        """
        self.__load_and_generate_output()

    def load_table(self, path):
        """some preprocessing on table

        1) load table, set timestamp as first column instead of index column
        2) drop duplicate rows based on column timestamp, keep first row
        3) reset index
        """
        table = pd.read_table(path, header=0, sep=",")
        first_column = table.columns.tolist()[0]
        table.drop_duplicates(subset=first_column, keep="first", inplace=True)
        table = table.reset_index(drop=True)
        return table, first_column

    def __load_and_generate_output(self):
        # metric table: timestamp,context-switches,branch-misses,......
        metrics, metric_first_column = self.load_table(self.metrics)
        # qos table:timestamp,qos
        qos, qos_first_column = self.load_table(self.qos)

        metrics_timestamps = metrics[metric_first_column].tolist()
        qos_timestamps = qos[qos_first_column].tolist()

        # get matched index list
        qos_filted_index, metrics_filted_index = self.__match_and_filter(
            qos_timestamps, metrics_timestamps
        )
        assert len(qos_filted_index) == len(metrics_filted_index)

        output_table = metrics.loc[metrics_filted_index]

        qos_filtered = qos.loc[qos_filted_index]
        qos_filtered.to_csv("qos.csv")
        col = qos_filtered.iloc[:, 1]

        output_table.insert(output_table.shape[1], "qos", col.values)
        output_table.to_csv(self.output, index=False)

    def __match_and_filter(self, qos_timestamps, metrics_timestamps):
        """get matched index
        Args:
           qos_timestamps: timestamp column of qos table
           metrics_timestamps: timestamp column of metric table
        """
        # should we expand qos table to get more data samples?
        qos_tss, metric_tss = [], []

        get_metrix_ts = lambda idx: int(
            time.mktime(time.strptime(metrics_timestamps[idx], "%Y-%m-%d %H:%M:%S"))
        )
        get_qos_ts = lambda idx: int(
            time.mktime(time.strptime(qos_timestamps[idx], "%Y-%m-%d %H:%M:%S"))
        )

        qos_index, metric_index = 0, 0
        # max time difference (in seconds) we can tolerate. AFAIAC 5 seconds is appropriate.
        max_diff = 5

        qos_ts = get_qos_ts(qos_index)
        metric_ts = get_metrix_ts(metric_index)

        while True:
            while abs(qos_ts - metric_ts) > max_diff:
                if qos_ts < metric_ts:
                    qos_index += 1
                else:
                    metric_index += 1
                if qos_index >= len(qos_timestamps) or metric_index >= len(
                    metrics_timestamps
                ):
                    break
                qos_ts = get_qos_ts(qos_index)
                metric_ts = get_metrix_ts(metric_index)

            if qos_index >= len(qos_timestamps) or metric_index >= len(
                metrics_timestamps
            ):
                break
            # append ideal index
            qos_tss.append(qos_index)
            metric_tss.append(metric_index)

            # increase index after append
            qos_index += 1
            metric_index += 1

        return qos_tss, metric_tss


class StressProcess(object):
    def __init__(self, stress, qos, output):
        self.stress = stress
        self.qos = qos
        self.qos_index = 0
        self.output = output

    def execute(self) -> None:
        """
        Execute Stress Data Process
        """
        self.__load_and_generate_output()

    def __load_and_generate_output(self):
        # begin-timestamp end-timestamp type stress command
        stress_col_name = ['begin-timestamp', 'end-timestamp', 'type', 'stress', 'command']
        stress_table = pd.read_table(self.stress, names=stress_col_name, header=0)
        # timestamp qos
        qos_col_name = ['timestamp', 'qos']
        qos_table = pd.read_table(self.qos, names=qos_col_name, header=0, seq=",")

        qos_len = len(qos_table)
        output_list = []

        # no_stress_qos = self.__get_rangetime_qos(None, stress_table.at[0, 'begin-timestamp'], qos_table)
        # output_list.append({"type": "none", "stress": "0", "avg-qos": no_stress_qos})

        for _, row in stress_table.iterrows():
            if self.qos_index >= qos_len:
                break

            begin_timestamp = row['begin-timestamp']
            end_timestamp = row['end-timestamp']
            average_qos = self.__get_rangetime_qos(begin_timestamp, end_timestamp, qos_table)
            output_list.append({"type": row['type'], "stress": row['stress'], "avg-qos": average_qos})

        # type stress avg-qos degradation-percent
        output_table = pd.DataFrame.from_records(output_list, columns=['type', 'stress', 'avg-qos', 'degradation-percent'])
        # output_table['degradation-percent'] = 100 * (output_table['avg-qos'] - no_stress_qos) / no_stress_qos

        no_stress_qos = -1
        for index, row in output_table.iterrows():
            if abs(row['stress']) <= 1e-5:
                no_stress_qos = row['avg-qos']
            row['degradation-percent'] = 100 * (row['avg-qos'] - no_stress_qos) / no_stress_qos
            output_table.iloc[index] = row

        output_table.to_csv(self.output, index=False)
                
    def __get_rangetime_qos(self, begin_time, end_time, qos_table):
        qos_len = len(qos_table)
        if self.qos_index >= qos_len:
            return 0

        if begin_time is not None:
            while self.__compare_stimestamp_gt(begin_time, qos_table.at[self.qos_index, 'timestamp']):
                self.qos_index += 1
                if self.qos_index >= qos_len:
                    return 0
        begin_index = self.qos_index

        while self.__compare_stimestamp_gt(end_time, qos_table.at[self.qos_index, 'timestamp']):
            self.qos_index += 1
            if self.qos_index >= qos_len:
                break
        end_index = self.qos_index

        return np.mean(qos_table[begin_index:end_index]["qos"])

    def __compare_stimestamp_gt(self, time1, time2):
        time1_st = int(time.mktime(time.strptime(
            time1, "%Y-%m-%d %H:%M:%S")))
        time2_st = int(time.mktime(time.strptime(
            time2, "%Y-%m-%d %H:%M:%S")))

        return time1_st > time2_st

class MachineProcess(object):
    def __init__(self, stress, machine, output):
        self.stress = stress
        self.machine = machine
        self.machine_index = 0
        self.output = output

    def execute(self) -> None:
        """
        Execute Cpu Usage Data Process
        """
        self.__load_and_generate_output()

    def __load_and_generate_output(self):
        # begin-timestamp end-timestamp type stress command
        stress_col_name = ['begin-timestamp', 'end-timestamp', 'type', 'stress', 'command']
        stress_table = pd.read_table(self.stress, names=stress_col_name, header=0)
        # timestamp cpu-usage memory cpi mkpi llc lmb rmb network-io disk-io
        machine_col_name = ['timestamp', 'cpu-usage', 'memory', 'cpi', 'mkpi', 'llc', 'lmb', 'rmb', 'network-io', 'disk-io']
        machine_table = pd.read_table(self.machine, names=machine_col_name, header=0)

        machine_len = len(machine_table)
        output_list = []

        for _, row in stress_table.iterrows():
            if self.machine_index >= machine_len:
                break

            begin_timestamp = row['begin-timestamp']
            end_timestamp = row['end-timestamp']
            average_cpu_usage = self.__get_rangetime_cpu_usage(begin_timestamp, end_timestamp, machine_table)
            output_list.append({"type": row['type'], "stress": row['stress'], "avg-cpu-usage": average_cpu_usage})

        # type stress avg-cpu-usage
        output_table = pd.DataFrame.from_records(output_list, columns=['type', 'stress', 'avg-cpu-usage'])
        output_table.to_csv(self.output, index=False)
                
    def __get_rangetime_cpu_usage(self, begin_time, end_time, machine_table):
        machine_len = len(machine_table)
        if self.machine_index >= machine_len:
            return 0

        if begin_time is not None:
            while self.__compare_stimestamp_gt(begin_time, machine_table.at[self.machine_index, 'timestamp']):
                self.machine_index += 1
                if self.machine_index >= machine_len:
                    return 0
        begin_index = self.machine_index

        while self.__compare_stimestamp_gt(end_time, machine_table.at[self.machine_index, 'timestamp']):
            self.machine_index += 1
            if self.machine_index >= machine_len:
                break
        end_index = self.machine_index

        return np.mean(machine_table[begin_index:end_index]["cpu-usage"])

    def __compare_stimestamp_gt(self, time1, time2):
        time1_st = int(time.mktime(time.strptime(
            time1, "%Y-%m-%d %H:%M:%S")))
        time2_st = int(time.mktime(time.strptime(
            time2, "%Y-%m-%d %H:%M:%S")))

        return time1_st > time2_st
