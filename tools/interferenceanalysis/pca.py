# %%
from sklearn import ensemble, neighbors, svm, linear_model, model_selection, tree, metrics, preprocessing 
from sklearn.cluster import KMeans
from sklearn.decomposition import PCA
from itertools import product
import streamlit as st
import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import altair as alt

def standardize_table(df):
    cols = list(df)
    for item in cols:
        if df[item].dtype == 'int64' or df[item].dtype == 'float64':
            mean_tmp = np.mean(np.array(df[item]))
            std_tmp = np.std(np.array(df[item]))
            if(std_tmp):
                df[item] = df[item].apply(lambda x: (x - mean_tmp) / std_tmp)
def normalize_table(df):
    cols = list(df)
    for item in cols:
        if df[item].dtype == 'int64' or df[item].dtype == 'float64':
            max_tmp = np.max(np.array(df[item]))
            min_tmp = np.min(np.array(df[item]))
            if (max_tmp != min_tmp):
                df[item] = df[item].apply(
                    lambda x: (x - min_tmp) * 100 / (max_tmp - min_tmp))
# %%
data = pd.read_csv("../tests/data/default/nginx.csv")
data_select = st.radio(
    """选择分析的数据集
    """,
    ('l3cache_stress.csv', 'nginx.csv'))
if data_select == 'l3cache_stress.csv':
    data = pd.read_csv("../tests/data/clickhouse/l3cache_stress.csv")
else:
    data=pd.read_csv("../tests/data/default/nginx.csv")
    data=data.drop(columns=['net','disk_io'])
mode = st.radio(
    "Please select a mode to visualize the data:",
    ('origin', 'normalization', 'standardization'))

if mode == 'normalization':
    normalize_table(data)
elif mode == 'standardization':
    standardize_table(data)

# %%

# %%
data_drop=data.drop(columns=['timestamp','qos'])
pca=PCA(n_components=6)
pca.fit(data_drop)
pca_predict=pca.transform(data_drop)

# %%
def plt_data(data):
    plt.xlabel('x label')
    plt.title('pca decomposition data')
    rev_data=pca_predict.T
    for i in range(len(rev_data)):
        plt.plot(rev_data[i])
    plt.show()
plt_data(pca_predict)

# %%
class PolynomialRegressionModel():
    """polynomial regression model for metrics"""

    def __init__(self, degree) -> None:
        self.model = linear_model.LinearRegression()

        self.mse = 0
        self.rmse = 0
        self.degree = degree

    def train_and_test_model(self):
        self.poly = preprocessing.PolynomialFeatures(degree = self.degree) 
        X_poly = self.poly.fit_transform(x_train)

        self.model.fit(X_poly, y_train)

        y_pred = self.model.predict(self.poly.transform(x_test))
        draw_comparison_altair_chart(y_test, y_pred)

        self.mse = metrics.mean_squared_error(y_test, y_pred)
        self.rmse = np.sqrt(metrics.mean_squared_error(y_test, y_pred))
        model_evaluation = "MSE: {}, RMSE: {}".format(self.mse, self.rmse)
        st.write(model_evaluation)


# %%
st.markdown("### 回归拟合分析")
# 数据预处理: 去除无效值; 特性缩放:标准化; 模型训练
x = pca_predict
print(x)
y = data[["qos"]]
x_train, x_test, y_train, y_test = model_selection.train_test_split(x, y, random_state=1)
def draw_comparison_altair_chart(y_test, y_pred):
    y_test_list = y_test["qos"]
    y_pred_list = y_pred if type(y_pred[0]) is not np.ndarray else [
        i[0] for i in y_pred]
    list_of_tuples = list(zip(y_test_list, y_pred_list))
    source = pd.DataFrame(list_of_tuples, columns=['Measured', 'Predicted'],
                          index=pd.RangeIndex(len(y_pred), name='index'))
    source = source.reset_index().melt('index', var_name='category', value_name='qos')
    nearest = alt.selection(type='single', nearest=True, on='mouseover',
                            fields=['index'], empty='none')
    line = alt.Chart(source).mark_line(interpolate='basis').encode(
        x='index:Q',
        y='qos:Q',
        color='category:N'
    )

    selectors = alt.Chart(source).mark_point().encode(
        x='index:Q',
        opacity=alt.value(0),
    ).add_selection(
        nearest
    )

    points = line.mark_point().encode(
        opacity=alt.condition(nearest, alt.value(1), alt.value(0))
    )

    # Draw text labels near the points, and highlight based on selection
    text = line.mark_text(align='left', dx=5, dy=-5).encode(
        text=alt.condition(nearest, 'qos:Q', alt.value(' '))
    )

    # Draw a rule at the location of the selection
    rules = alt.Chart(source).mark_rule(color='gray').encode(
        x='index:Q',
    ).transform_filter(
        nearest
    )

    charts = alt.layer(
        line, selectors, points, rules, text
    ).interactive()
    st.altair_chart(charts, use_container_width=True)


# %%

class RegressionModel():
    """regression model for metrics"""

    def __init__(self, model) -> None:
        self.model = model

        self.mse = 0
        self.rmse = 0

    def train_and_test_model(self):
        self.model.fit(x_train, np.ravel(y_train))

        y_pred = self.model.predict(x_test)
        draw_comparison_altair_chart(y_test, y_pred)

        self.mse = metrics.mean_squared_error(y_test, y_pred)
        self.rmse = np.sqrt(metrics.mean_squared_error(y_test, y_pred))
        model_evaluation = "MSE: {}, RMSE: {}".format(self.mse, self.rmse)
        st.write(model_evaluation)


# %%
polynomial_regression_list = [
    {'name': '1st Degree Polynomial Regression', 'degree': 1},
    {'name': '2nd Degree Polynomial Regression', 'degree': 2},
    {'name': '3rd Degree Polynomial Regression', 'degree': 3},
    {'name': '4th Degree Polynomial Regression', 'degree': 4},
    {'name': '5th Degree Polynomial Regression', 'degree': 5},
]

other_regression_list = [
    {'name': 'Decision Tree Regression', 'regressor': tree.DecisionTreeRegressor()},
    {'name': 'SVM Regression', 'regressor': svm.SVR()},
    {'name': 'KNN Regression', 'regressor': neighbors.KNeighborsRegressor()},
    {'name': 'Random Forest Regression', 'regressor': ensemble.RandomForestRegressor(n_estimators=20)},
    {'name': 'Adaboost Regression', 'regressor': ensemble.AdaBoostRegressor(n_estimators=50)},
    {'name': 'Gradient Boosting Regression', 'regressor': ensemble.GradientBoostingRegressor(n_estimators=100)},
    {'name': 'Bagging Regression', 'regressor': ensemble.BaggingRegressor()},
    {'name': 'ExtraTree Regression', 'regressor': tree.ExtraTreeRegressor()},
    {'name': 'ElasticNet','regressor':linear_model.ElasticNet()},
]

mse_list = []
polynomial_list = []


# %%
for polynomial_regression in polynomial_regression_list:
    st.markdown("#### " + polynomial_regression['name'])
    degree = polynomial_regression['degree']
    
    regression_model = PolynomialRegressionModel(degree)
    regression_model.train_and_test_model()

    polynomial_list.append({'name': polynomial_regression['name'], 'rmse': regression_model.rmse, 'regression_model': regression_model})
    mse_list.append({'name': polynomial_regression['name'], 'mse': regression_model.mse, 'rmse': regression_model.rmse})

for other_regression in other_regression_list:
    st.markdown("#### " + other_regression['name'])
    regressor = other_regression['regressor']
    
    regression_model = RegressionModel(regressor)
    regression_model.train_and_test_model()

    mse_list.append({'name': other_regression['name'], 'mse': regression_model.mse, 'rmse': regression_model.rmse})

# %%
def takeMse(elem):
    return elem['rmse']
mse_list.sort(key=takeMse)
st.table(pd.DataFrame(mse_list).set_index('name'))


