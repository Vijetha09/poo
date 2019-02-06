
name := "Vmmetrics"

version := "0.1"

scalaVersion := "2.11.12"


libraryDependencies += "org.apache.spark" % "spark-streaming_2.11" % "2.3.0"

libraryDependencies += "org.apache.spark" % "spark-streaming-kafka-0-10_2.11" % "2.3.0"

libraryDependencies += "org.apache.logging.log4j" % "log4j-api" % "2.11.0"

libraryDependencies += "org.apache.logging.log4j" % "log4j-core" % "2.11.0"

libraryDependencies += "org.apache.spark" % "spark-sql_2.11" % "2.3.0"

libraryDependencies += "org.apache.spark" % "spark-sql-kafka-0-10_2.11" % "2.3.0"


libraryDependencies += "com.fasterxml.jackson.core" % "jackson-core" % "2.8.7"
libraryDependencies += "com.fasterxml.jackson.core" % "jackson-databind" % "2.8.7"
libraryDependencies += "com.fasterxml.jackson.module" % "jackson-module-scala_2.11" % "2.8.7"

libraryDependencies += "org.mongodb.spark" %% "mongo-spark-connector" % "2.2.0"

libraryDependencies +=  "com.typesafe" % "config" % "1.3.2"

libraryDependencies +=  "net.jpountz.lz4" % "lz4" % "1.3.0"

assemblyMergeStrategy in assembly := {
  case PathList("META-INF", xs @ _*) => MergeStrategy.discard
  case x => MergeStrategy.first
}
