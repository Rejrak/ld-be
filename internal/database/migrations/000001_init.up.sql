CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    kc_id UUID,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    nickname TEXT NOT NULL,
    admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS training_plan (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS workout (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    training_plan_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    FOREIGN KEY (training_plan_id) REFERENCES training_plan(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS exercise (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    workout_id UUID NOT NULL,
    exercise_type_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    FOREIGN KEY (exercise_type_id) REFERENCES exercise_type(id) ON DELETE CASCADE,
    FOREIGN KEY (workout_id) REFERENCES workout(id) ON DELETE CASCADE
);

CREATE TABLE IF NO EXISTS exercise_type (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);
CREATE TABLE IF NOT EXISTS exercise_set (
    id UUID PRIMARY KEY,
    exercise_id UUID NOT NULL,
    weight DECIMAL(10, 2) NOT NULL,
    reps INT NOT NULL,
    rest_time INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    FOREIGN KEY (exercise_id) REFERENCES exercise(id) ON DELETE CASCADE
);
