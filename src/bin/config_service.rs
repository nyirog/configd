use std::{error::Error, future::pending};
use zbus::{connection, interface};
use zvariant::{Type, Structure, StructureBuilder};
use serde::{Serialize, Deserialize};

#[derive(Serialize, Deserialize, Type, Debug, Clone, PartialEq)]
struct Config {
    some_int: i64,
    some_string: String,
}

struct ConfigMethod {
    config: Config,
    applied: bool,
}

impl From<Config> for Structure<'_> {
    fn from(value: Config) -> Self {
        StructureBuilder::new()
            .add_field(value.some_int)
            .add_field(value.some_string)
            .build().unwrap()
    }
}

#[interface(name = "org.configd.Config")]
impl ConfigMethod {
    fn apply(&mut self) -> zbus::fdo::Result<()> {
        self.applied = true;
        Ok(())
    }

    #[zbus(property)]
    fn config(&self) -> Config {
        self.config.clone()
    }

    #[zbus(property)]
    fn set_config(&mut self, config: (i64, String)) {
        self.config = Config {
            some_int: config.0,
            some_string: config.1,
        };
    }
}

// Although we use `tokio` here, you can use any async runtime of choice.
#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    let config = ConfigMethod {
        config: Config {
            some_int: 42,
            some_string: "hello".to_string(),
        },
        applied: true,
    };
    let _conn = connection::Builder::session()?
        .name("org.configd")?
        .serve_at("/org/configd", config)?
        .build()
        .await?;

    // Do other things or go to wait forever
    pending::<()>().await;

    Ok(())
}
