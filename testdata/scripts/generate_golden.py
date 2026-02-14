#!/usr/bin/env python3
"""Generate golden test data for go-lgbm.

Trains small LightGBM models for each objective type, saves text-format
model files and JSON reference predictions. Generates models compatible
with LightGBM v3 and v4 text formats.

Requirements:
    pip install lightgbm numpy

Usage:
    python generate_golden.py
"""

import json
import os
import sys

import lightgbm as lgb
import numpy as np

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
TESTDATA_DIR = os.path.dirname(SCRIPT_DIR)

# Fixed seed for reproducibility
RNG = np.random.RandomState(42)

# Small dataset size for fast tests
N_TRAIN = 200
N_TEST = 5
N_FEATURES = 10


def generate_binary(output_dir: str) -> None:
    """Generate binary classification model and reference predictions."""
    X_train = RNG.randn(N_TRAIN, N_FEATURES)
    y_train = (X_train[:, 0] + X_train[:, 1] > 0).astype(int)

    X_test = RNG.randn(N_TEST, N_FEATURES)

    params = {
        "objective": "binary",
        "num_leaves": 8,
        "learning_rate": 0.1,
        "n_estimators": 20,
        "verbose": -1,
    }

    model = lgb.LGBMClassifier(**params)
    model.fit(X_train, y_train)

    model_path = os.path.join(output_dir, "binary.txt")
    model.booster_.save_model(model_path)

    preds = model.predict_proba(X_test)[:, 1].tolist()
    raw_preds = model.predict_proba(X_test, raw_score=True).tolist()

    ref = {
        "inputs": X_test.tolist(),
        "predictions": preds,
        "raw_predictions": raw_preds,
    }

    ref_path = os.path.join(output_dir, "binary.json")
    with open(ref_path, "w") as f:
        json.dump(ref, f, indent=2)

    print(f"  binary: model={model_path}, ref={ref_path}")


def generate_regression(output_dir: str) -> None:
    """Generate regression model and reference predictions."""
    X_train = RNG.randn(N_TRAIN, N_FEATURES)
    y_train = X_train[:, 0] * 2.0 + X_train[:, 1] + RNG.randn(N_TRAIN) * 0.1

    X_test = RNG.randn(N_TEST, N_FEATURES)

    params = {
        "objective": "regression",
        "num_leaves": 8,
        "learning_rate": 0.1,
        "n_estimators": 20,
        "verbose": -1,
    }

    model = lgb.LGBMRegressor(**params)
    model.fit(X_train, y_train)

    model_path = os.path.join(output_dir, "regression.txt")
    model.booster_.save_model(model_path)

    preds = model.predict(X_test).tolist()

    ref = {
        "inputs": X_test.tolist(),
        "predictions": preds,
    }

    ref_path = os.path.join(output_dir, "regression.json")
    with open(ref_path, "w") as f:
        json.dump(ref, f, indent=2)

    print(f"  regression: model={model_path}, ref={ref_path}")


def generate_multiclass(output_dir: str) -> None:
    """Generate multiclass classification model and reference predictions."""
    X_train = RNG.randn(N_TRAIN, N_FEATURES)
    y_train = (X_train[:, 0] > 0.5).astype(int) + (X_train[:, 1] > 0).astype(int)

    X_test = RNG.randn(N_TEST, N_FEATURES)

    params = {
        "objective": "multiclass",
        "num_class": 3,
        "num_leaves": 8,
        "learning_rate": 0.1,
        "n_estimators": 20,
        "verbose": -1,
    }

    model = lgb.LGBMClassifier(**params)
    model.fit(X_train, y_train)

    model_path = os.path.join(output_dir, "multiclass.txt")
    model.booster_.save_model(model_path)

    preds = model.predict_proba(X_test).tolist()

    ref = {
        "inputs": X_test.tolist(),
        "predictions": preds,
    }

    ref_path = os.path.join(output_dir, "multiclass.json")
    with open(ref_path, "w") as f:
        json.dump(ref, f, indent=2)

    print(f"  multiclass: model={model_path}, ref={ref_path}")


def generate_ranking(output_dir: str) -> None:
    """Generate ranking model and reference predictions."""
    X_train = RNG.randn(N_TRAIN, N_FEATURES)
    y_train = RNG.randint(0, 5, N_TRAIN).astype(float)
    # 4 groups of 50
    group_train = [50, 50, 50, 50]

    X_test = RNG.randn(N_TEST, N_FEATURES)

    train_data = lgb.Dataset(X_train, label=y_train, group=group_train)

    params = {
        "objective": "lambdarank",
        "num_leaves": 8,
        "learning_rate": 0.1,
        "verbose": -1,
    }

    model = lgb.train(params, train_data, num_boost_round=20)

    model_path = os.path.join(output_dir, "ranking.txt")
    model.save_model(model_path)

    preds = model.predict(X_test).tolist()

    ref = {
        "inputs": X_test.tolist(),
        "predictions": preds,
    }

    ref_path = os.path.join(output_dir, "ranking.json")
    with open(ref_path, "w") as f:
        json.dump(ref, f, indent=2)

    print(f"  ranking: model={model_path}, ref={ref_path}")


def main() -> None:
    lgb_version = lgb.__version__
    major = int(lgb_version.split(".")[0])
    print(f"LightGBM version: {lgb_version} (major={major})")

    if major == 3:
        output_dir = os.path.join(TESTDATA_DIR, "v3")
    elif major == 4:
        output_dir = os.path.join(TESTDATA_DIR, "v4")
    else:
        print(f"ERROR: Unsupported LightGBM version {lgb_version}")
        sys.exit(1)

    os.makedirs(output_dir, exist_ok=True)
    print(f"Output directory: {output_dir}")

    generate_binary(output_dir)
    generate_regression(output_dir)
    generate_multiclass(output_dir)
    generate_ranking(output_dir)

    print(f"\nDone. To generate for the other version, install a different")
    print(f"LightGBM major version and re-run this script.")


if __name__ == "__main__":
    main()
