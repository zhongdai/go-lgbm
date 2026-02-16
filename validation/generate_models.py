"""Generate LightGBM models and test inputs for validation.

Trains 4 model types (binary, multiclass, regression, ranking) using
synthetic datasets with a fixed random seed, exports each as LightGBM
v3 text format (for leaves compatibility), and generates 1,000 random
test inputs per model as JSON.
"""

import json
import os

import lightgbm as lgb
import numpy as np
from sklearn.datasets import make_classification, make_regression

SEED = 42
N_FEATURES = 10
N_TRAIN = 2000
N_TEST = 1000
MODELS_DIR = "models"
TESTDATA_DIR = "testdata"


def ensure_dirs():
    os.makedirs(MODELS_DIR, exist_ok=True)
    os.makedirs(TESTDATA_DIR, exist_ok=True)


def save_model_v3(model, path):
    """Save model in v3 text format (compatible with leaves library).

    LightGBM 4.x saves v4 format by default. The leaves library only
    supports v3. Converting is safe: replace the version header and
    remove the tree_sizes line (a v4 addition).
    """
    model_str = model.model_to_string()
    lines = model_str.split("\n")
    converted = []
    for line in lines:
        if line == "version=v4":
            converted.append("version=v3")
        else:
            converted.append(line)
    with open(path, "w") as f:
        f.write("\n".join(converted))


def generate_binary():
    """Train a binary classification model."""
    X, y = make_classification(
        n_samples=N_TRAIN,
        n_features=N_FEATURES,
        n_informative=6,
        n_redundant=2,
        random_state=SEED,
    )
    ds = lgb.Dataset(X, label=y)
    params = {
        "objective": "binary",
        "metric": "binary_logloss",
        "num_leaves": 31,
        "learning_rate": 0.1,
        "verbose": -1,
        "seed": SEED,
    }
    model = lgb.train(params, ds, num_boost_round=50)
    save_model_v3(model, os.path.join(MODELS_DIR, "binary.txt"))

    rng = np.random.RandomState(SEED)
    inputs = rng.randn(N_TEST, N_FEATURES).tolist()
    with open(os.path.join(TESTDATA_DIR, "binary.json"), "w") as f:
        json.dump({"inputs": inputs, "n_features": N_FEATURES}, f)
    print(f"  binary: {N_TRAIN} train, {N_TEST} test inputs, {N_FEATURES} features")


def generate_multiclass():
    """Train a multiclass classification model."""
    n_classes = 5
    X, y = make_classification(
        n_samples=N_TRAIN,
        n_features=N_FEATURES,
        n_informative=6,
        n_redundant=2,
        n_classes=n_classes,
        n_clusters_per_class=1,
        random_state=SEED,
    )
    ds = lgb.Dataset(X, label=y)
    params = {
        "objective": "multiclass",
        "num_class": n_classes,
        "metric": "multi_logloss",
        "num_leaves": 31,
        "learning_rate": 0.1,
        "verbose": -1,
        "seed": SEED,
    }
    model = lgb.train(params, ds, num_boost_round=50)
    save_model_v3(model, os.path.join(MODELS_DIR, "multiclass.txt"))

    rng = np.random.RandomState(SEED + 1)
    inputs = rng.randn(N_TEST, N_FEATURES).tolist()
    with open(os.path.join(TESTDATA_DIR, "multiclass.json"), "w") as f:
        json.dump(
            {"inputs": inputs, "n_features": N_FEATURES, "n_classes": n_classes}, f
        )
    print(f"  multiclass: {N_TRAIN} train, {N_TEST} test inputs, {n_classes} classes")


def generate_regression():
    """Train a regression model."""
    X, y = make_regression(
        n_samples=N_TRAIN,
        n_features=N_FEATURES,
        n_informative=6,
        noise=0.1,
        random_state=SEED,
    )
    ds = lgb.Dataset(X, label=y)
    params = {
        "objective": "regression",
        "metric": "rmse",
        "num_leaves": 31,
        "learning_rate": 0.1,
        "verbose": -1,
        "seed": SEED,
    }
    model = lgb.train(params, ds, num_boost_round=50)
    save_model_v3(model, os.path.join(MODELS_DIR, "regression.txt"))

    rng = np.random.RandomState(SEED + 2)
    inputs = rng.randn(N_TEST, N_FEATURES).tolist()
    with open(os.path.join(TESTDATA_DIR, "regression.json"), "w") as f:
        json.dump({"inputs": inputs, "n_features": N_FEATURES}, f)
    print(f"  regression: {N_TRAIN} train, {N_TEST} test inputs")


def generate_ranking():
    """Train a ranking (lambdarank) model."""
    X, y_raw = make_regression(
        n_samples=N_TRAIN,
        n_features=N_FEATURES,
        n_informative=6,
        noise=0.1,
        random_state=SEED + 3,
    )
    # Convert to relevance labels 0-4
    y = np.clip(
        np.round((y_raw - y_raw.min()) / (y_raw.max() - y_raw.min()) * 4), 0, 4
    ).astype(int)
    # Create groups of 20 documents each
    group_size = 20
    n_groups = N_TRAIN // group_size
    groups = [group_size] * n_groups

    ds = lgb.Dataset(X[: n_groups * group_size], label=y[: n_groups * group_size])
    ds.set_group(groups)
    params = {
        "objective": "lambdarank",
        "metric": "ndcg",
        "num_leaves": 31,
        "learning_rate": 0.1,
        "verbose": -1,
        "seed": SEED,
    }
    model = lgb.train(params, ds, num_boost_round=50)
    save_model_v3(model, os.path.join(MODELS_DIR, "ranking.txt"))

    rng = np.random.RandomState(SEED + 4)
    inputs = rng.randn(N_TEST, N_FEATURES).tolist()
    with open(os.path.join(TESTDATA_DIR, "ranking.json"), "w") as f:
        json.dump({"inputs": inputs, "n_features": N_FEATURES}, f)
    print(f"  ranking: {N_TRAIN} train, {N_TEST} test inputs, {n_groups} groups")


def main():
    ensure_dirs()
    print("Generating models and test data...")
    generate_binary()
    generate_multiclass()
    generate_regression()
    generate_ranking()
    print("Done.")


if __name__ == "__main__":
    main()
